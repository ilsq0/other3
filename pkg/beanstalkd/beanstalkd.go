package beanstalkd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"

	"sync/atomic"
	"time"
)

var (
	ErrBk  = errors.New("ErrBk")
	ErrNet = errors.New("ErrNet")
)

type Conn struct {
	tc      *textproto.Conn
	lastUse time.Time
	using   string
	err     error
}

func Dial(ctx context.Context, addr string) (*Conn, error) {
	d := &net.Dialer{
		Timeout: time.Second * 3,
	}
	netConn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %w: %v", ErrNet, err)
	}
	conn := &Conn{
		tc:      textproto.NewConn(netConn),
		lastUse: time.Now(),
	}
	return conn, nil
}

func (c *Conn) Use(tube string) error {
	if c.using == tube {
		return nil
	}
	c.err = c.tc.PrintfLine("use %s", tube)
	if c.err != nil {
		c.err = fmt.Errorf("use %w: %v", ErrNet, c.err)
		return c.err
	}
	resp, err := c.tc.ReadLine()
	if err != nil {
		c.err = fmt.Errorf("use %w: %v", ErrNet, err)
		return c.err
	}
	if _, err := fmt.Sscanf(resp, "USING %s", &tube); err != nil {
		c.err = fmt.Errorf("use %w: %v", ErrBk, resp)
	} else {
		c.using = tube
	}
	return c.err
}

func (c *Conn) Watch(tube string) (uint64, error) {
	c.err = c.tc.PrintfLine("watch " + tube)
	if c.err != nil {
		c.err = fmt.Errorf("watch %w: %v", ErrNet, c.err)
		return 0, c.err
	}
	resp, err := c.tc.ReadLine()
	if err != nil {
		c.err = fmt.Errorf("watch %w: %v", ErrNet, err)
		return 0, c.err
	}
	var id uint64
	if _, err := fmt.Sscanf(resp, "WATCHING %d", &id); err != nil {
		c.err = fmt.Errorf("watch %w: %v", ErrBk, resp)
	}
	return id, c.err
}

var _crnl = []byte{'\r', '\n'}

func (c *Conn) Put(tube string, body []byte, pri int, delay, ttr time.Duration) (uint64, error) {
	if err := c.Use(tube); err != nil {
		return 0, err
	}
	fmt.Fprintf(c.tc.W, "put %d %d %d %d\r\n", pri, int(delay.Seconds()), int(ttr.Seconds()), len(body))
	c.tc.W.Write(body)
	c.tc.W.Write(_crnl)
	if c.err = c.tc.W.Flush(); c.err != nil {
		c.err = fmt.Errorf("put %w: %v", ErrNet, c.err)
		return 0, c.err
	}
	resp, err := c.tc.ReadLine()
	if err != nil {
		c.err = fmt.Errorf("put %w: %v", ErrNet, err)
		return 0, c.err
	}
	var id uint64
	if _, err := fmt.Sscanf(resp, "INSERTED %d", &id); err != nil {
		c.err = fmt.Errorf("put %w: %v", ErrBk, resp)
	}
	return id, c.err
}

func (c *Conn) Delete(id uint64) error {
	c.err = c.tc.PrintfLine("delete %d", id)
	if c.err != nil {
		c.err = fmt.Errorf("delete %w: %v", ErrNet, c.err)
		return c.err
	}
	resp, err := c.tc.ReadLine()
	if err != nil {
		c.err = fmt.Errorf("delete %w: %v", ErrNet, err)
		return c.err
	}
	if resp != "DELETED" {
		c.err = fmt.Errorf("delete %w: %v", ErrBk, resp)
	}
	return c.err
}

func (c *Conn) ReserveWithTimeout(seconds int) (uint64, []byte, error) {
	c.err = c.tc.PrintfLine(`reserve-with-timeout %d`, seconds)
	if c.err != nil {
		c.err = fmt.Errorf("reserve %w: %v", ErrNet, c.err)
		return 0, nil, c.err
	}
	resp, err := c.tc.ReadLine()
	if err != nil {
		c.err = fmt.Errorf("reserve %w: %v", ErrNet, err)
		return 0, nil, c.err
	}
	var id, len uint64
	if _, err := fmt.Sscanf(resp, "RESERVED %d %d", &id, &len); err != nil {
		c.err = fmt.Errorf("reserve %w: %v", ErrBk, resp)
		return 0, nil, c.err
	}
	body := make([]byte, len+2)
	if _, err := io.ReadFull(c.tc.R, body); err != nil {
		c.err = fmt.Errorf("reserve %w: %v", ErrNet, err)
		return 0, nil, c.err
	}
	return id, body[:len], c.err
}

func (c *Conn) Close() error {
	return c.tc.Close()
}

type Pool struct {
	factory func(context.Context) (*Conn, error)
	active  atomic.Int64
	ch      chan *Conn
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewPool(pctx context.Context, cap int, idleTimeout time.Duration, addr string) *Pool {
	p := &Pool{
		factory: func(ctx context.Context) (*Conn, error) { return Dial(ctx, addr) },
		ch:      make(chan *Conn, cap),
	}
	p.ctx, p.cancel = context.WithCancel(pctx)
	for i := 0; i < cap; i++ {
		p.ch <- nil
	}
	if idleTimeout > 0 {
		go p.loopIdle(idleTimeout, cap)
	}
	return p
}

func (p *Pool) GetWithTimeout(timeout time.Duration) (*Conn, error) {
	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-p.ch:
		if conn == nil {
			p.active.Add(1)
			newConn, err := p.factory(ctx)
			if err != nil {
				p.active.Add(-1)
				p.ch <- nil
				return nil, err
			}
			return newConn, nil
		}
		return conn, nil
	}
}

func (p *Pool) Put(conn *Conn) {
	select {
	case <-p.ctx.Done():
		conn.Close()
		p.active.Add(-1)
		return
	default:
	}
	if errors.Is(conn.err, ErrNet) {
		conn.Close()
		conn = nil
		p.active.Add(-1)
	} else {
		conn.err = nil
		conn.lastUse = time.Now()
	}
	p.ch <- conn
}

func (p *Pool) loopIdle(idleTimeout time.Duration, cap int) {
	timer := time.NewTimer(idleTimeout / 10)
	defer timer.Stop()
	for {
		select {
		case <-p.ctx.Done():
			p.clean()
			return
		case <-timer.C:
			for i := 0; i < cap; i++ {
				select {
				case conn := <-p.ch:
					if conn != nil && time.Since(conn.lastUse) > idleTimeout {
						conn.Close()
						conn = nil
						p.active.Add(-1)
					}
					p.ch <- conn
				default:
					goto forOut
				}
			}
		forOut:
			timer.Reset(idleTimeout / 10)
		}
	}
}

func (p *Pool) clean() {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case conn := <-p.ch:
			if conn != nil {
				conn.Close()
				p.active.Add(-1)
			}
		case <-timer.C:
			if p.active.Load() == 0 {
				return
			}
			timer.Reset(time.Second)
		}
	}
}

func (p *Pool) Stat() string {
	return fmt.Sprintf("active=%d", p.active.Load())
}
