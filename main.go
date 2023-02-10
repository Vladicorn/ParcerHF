package main

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type CodeError struct {
	Code        string
	Description string
}

func ParseSite() {
	dataTable := make([]CodeError, 0, 6)

	res, err := http.Get("https://confluence.hflabs.ru/pages/viewpage.action?pageId=1181220999")
	checkError(err)

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		data := CodeError{}
		s.Find("td").Each(func(i1 int, s2 *goquery.Selection) {
			if i1 == 0 {
				data.Code = s2.Text()
			}
			if i1 == 1 {
				data.Description = s2.Text()
			}
		})
		dataTable = append(dataTable, data)
	})
	GoogleSheetWrite(dataTable)

}

func GoogleSheetWrite(dataTable []CodeError) {
	data, err := ioutil.ReadFile("client_secret.json")
	checkError(err)

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	checkError(err)

	client := conf.Client(context.TODO())

	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := service.FetchSpreadsheet("1OePl8qPJWD3IOHwtkiJT9PBBLryM0Ifu5MKS399ijyw")
	checkError(err)
	sheet, err := spreadsheet.SheetByIndex(0)
	checkError(err)

	for index, colum := range dataTable {
		if sheet.Rows[index][0].Value != colum.Code {
			sheet.Update(index, 0, colum.Code)
		}
		if sheet.Rows[index][1].Value != colum.Description {
			sheet.Update(index, 1, colum.Description)
		}
	}

	err = sheet.Synchronize()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ParseSite()
}
