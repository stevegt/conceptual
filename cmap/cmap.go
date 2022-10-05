package cmap

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	. "github.com/stevegt/goadapt"
)

const (
	DUPNOK int = iota
	DUPSKIP
)

type Node struct {
	Name  string
	Label string
}

type Phrase struct {
	Src      string
	Relation string
	Dst      string
}

// map[nodeName]*Node
type Nodes map[string]*Node

type CMap struct {
	Cmdline string
	Name    string
	Title   string
	Txt     string
	Nodes   Nodes
	// maintain ordered lists so we can provide reproducible output
	NodeNames []string
	Phrases   []*Phrase
}

func Load(fh io.Reader, cmdline string) (cm *CMap, err error) {
	defer Return(&err)
	cm = &CMap{}
	cm.Cmdline = cmdline
	cm.Nodes = make(Nodes)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		txt := scanner.Text()
		cm.AddRule(txt)
		cm.Txt += Spf("%s\n", txt)
	}
	err = scanner.Err()
	Ck(err)

	// cm.Verify()

	return
}

type CMErr struct {
	msg string
	Rc  int
}

func (e CMErr) Error() string {
	return e.msg
}

func CMErrIf(cond bool, rc int, args ...interface{}) {
	if cond {
		msg := fmt.Sprintf(args[0].(string), args[1:]...)
		panic(CMErr{msg: msg, Rc: rc})
	}
	return
}

func (cm *CMap) EnsureNode(name, label string, flag int) {
	_, ok := cm.Nodes[name]
	CMErrIf(ok && flag == DUPNOK, 3, "duplicate node name: %s", name)
	if ok {
		return
	}

	n := &Node{
		Name:  name,
		Label: label,
	}
	cm.Nodes[name] = n
	cm.NodeNames = appendUniq(cm.NodeNames, name)
}

func (cm *CMap) AddRule(txt string) {
	parts := strings.Fields(txt)
	if len(parts) == 0 {
		return
	}
	typ := parts[0]
	switch typ {
	case "//":
		return
	case "name":
		cm.Name = parts[1]
	case "title":
		cm.Title = strings.Join(parts[1:], " ")
	case "a":
		CMErrIf(len(parts) < 3, 2, "missing node label: %s", txt)
		name := parts[1]
		label := strings.Join(parts[2:], " ")

		cm.EnsureNode(name, label, DUPNOK)

	case "p":
		CMErrIf(len(parts) < 2, 4, "missing phrase src: %s", txt)
		CMErrIf(len(parts) < 3, 5, "missing phrase relation: %s", txt)
		CMErrIf(len(parts) < 4, 6, "missing phrase dst: %s", txt)

		srcname := parts[1]
		relation := strings.Join(parts[2:len(parts)-1], " ")
		dstname := parts[len(parts)-1]

		cm.EnsureNode(srcname, srcname, DUPSKIP)
		cm.EnsureNode(dstname, dstname, DUPSKIP)

		p := &Phrase{
			Src:      srcname,
			Relation: relation,
			Dst:      dstname,
		}

		cm.Phrases = append(cm.Phrases, p)
	default:
		CMErrIf(true, 12, "unrecognized rule: %s", txt)
	}
}

/*
func (cm *CMap) Verify() {
	for stateName, state := range m.States {
		for _, eventName := range m.EventNames {
			_, ok := state.Transitions[eventName]
			CMErrIf(!ok, 13, "unhandled event: machine %v, state %v, event %v", m.Package, stateName, eventName)
		}
	}
}
*/

func appendUniq(in []string, add string) (out []string) {
	out = in[:]
	if add == "" {
		return
	}
	found := false
	for _, s := range out {
		if s == add {
			found = true
			break
		}
	}
	if !found {
		out = append(out, add)
	}
	return
}

func (cm *CMap) LsNodes() (out []*Node) {
	for _, name := range cm.NodeNames {
		out = append(out, cm.Nodes[name])
	}
	return
}

func (cm *CMap) LsPhrases() (out []*Phrase) {
	for _, p := range cm.Phrases {
		out = append(out, p)
	}
	return
}

/*
func (m *Machine) LsEvents() (out []string) {
	for _, srcName := range m.stateNames {
		src := m.States[srcName]
		for _, event := range src.events {
			found := false
			for _, e := range out {
				if event == e {
					found = true
					break
				}
			}
			if !found {
				out = append(out, event)
			}
		}
	}
	return
}

func (m *Machine) LsMethods() (out []string) {
	for _, srcName := range m.stateNames {
		src := m.States[srcName]
		for _, event := range src.events {
			found := false
			for _, e := range out {
				if event == e {
					found = true
					break
				}
			}
			if !found {
				out = append(out, event)
			}
		}
	}
	return
}
*/

//go:embed template/*
var fs embed.FS

func (cm *CMap) ToDot() (out []byte) {
	t := template.Must(template.ParseFS(fs, "template/dot.ttmpl"))
	var buf bytes.Buffer
	err := t.Execute(&buf, cm)
	Ck(err)
	out = buf.Bytes()
	return
}

/*
func (cm *CMap) ToGo() (out []byte) {
	t := template.Must(template.ParseFS(fs, "template/go.ttmpl"))
	var buf bytes.Buffer
	err := t.Execute(&buf, m)
	Ck(err)
	out = buf.Bytes()
	return
}
*/
