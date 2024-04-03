package srvgenerator

import (
	"fmt"
	"html/template"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"time"
)

type data struct {
	SRV    string
	Serial int
	Host   string
	Ip     string
	Port   int
}

const tpl = `
$ORIGIN {{.SRV}}.
;; SOA Record
@	3600 IN SOA ns1.yx. admin.yx. (
           {{.Serial}}  ; Serial number
           7200         ; Refresh time (2 hours)
           3600         ; Retry time (1 hour)
           1209600      ; Expiry time (2 weeks)
           3600         ; Minimum TTL (1 hour)
        )

;; A Records
_ip IN A {{.Ip}}

;; SRV Records
@ IN SRV 1 90 {{.Port}} _ip.{{.SRV}}.
`

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NotifyCoredns(srv string, port int, script string) error {
	fn, err := renderTemplate(srv, port)
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", script, fn)
	return cmd.Run()
}

func renderTemplate(srv string, port int) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		return "", err
	}
	addrs, err := eth0.Addrs()
	if err != nil {
		return "", err
	}
	ip4 := ""
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ip4 = ipNet.IP.String()
			break
		}
	}
	if ip4 == "" {
		return "", fmt.Errorf("there is no ip4 on eth0")
	}
	data := data{
		SRV:    srv,
		Serial: r.Intn(99999999),
		Host:   host,
		Ip:     ip4,
		Port:   port,
	}
	tmpl, err := template.New("tmpl").Parse(tpl)
	if err != nil {
		return "", err
	}
	// generate the db.domain file
	fn := fmt.Sprintf("db.%s", srv)
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return fn, tmpl.Execute(f, data)
}
