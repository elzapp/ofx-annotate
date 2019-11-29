package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	ofx "github.com/elzapp/go-ofx"
)

type ptrn struct {
	Name     string `json:"name"`
	Patterns []string `json:"patterns"`
}

func newPtrn(name string, patterns ...string) ptrn {
	return ptrn{name, patterns}
}

func main() {
	var source string = "-"
	if len(os.Args) >= 2 {
		source = os.Args[1]
	}
	var txs ofx.OfxTransactionList
	var ptrns []ptrn
	j,_ := os.Open("patterns.json")
	js,_ := ioutil.ReadAll(j)
	json.Unmarshal(js, &ptrns)
	//js,_ = json.MarshalIndent(ptrns,"","  ")
	//fmt.Fprintln(os.Stderr,string(js))
	reader := os.Stdin
	if source != "-" {
		file, _ := os.Open(source)
		reader = file
	}
	txdata, _ := ioutil.ReadAll(reader)
	if source != "-" {
		reader.Close()
	}
	xml.Unmarshal(txdata, &txs)
	re := regexp.MustCompile(`([0-9]{2}\.[0-9]{2}|\*[0-9]{4} [0-9]+.[0-9]+ [A-Z]{3} [0-9]+.[0-9]+|[A-Z]{3} [.0-9]+|) ?(.*)`)
	for txi, tx := range txs.Transactions {
		m := re.FindStringSubmatch(tx.Memo)
		//fmt.Println(m[0:len(m)-1])
		
		normalized := m[len(m)-1]
		fmt.Fprintf(os.Stderr,"%s\n",normalized)
		var found bool
		for _, p := range ptrns {
			for _, rs := range p.Patterns {
				pr := regexp.MustCompile(rs)
				if pr.MatchString(normalized) {
					tx.Payee = p.Name
					fmt.Fprintf(os.Stderr,"--> %s\n",p.Name)
					txs.Transactions[txi] = tx
					found = true
				}
			}
		}
		if !found {
			fmt.Println("!!! "+normalized)
		}

	}
	output,_ := xml.MarshalIndent(txs,"","  ")
	if source == "-" {
		fmt.Println(string(output))
	} else {
		f,_ := os.Create(source)
		f.Write(output)
		f.Close()
	}
}
