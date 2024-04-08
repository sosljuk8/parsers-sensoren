package main

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/gocolly/colly/v2"
	"hash/fnv"
	"log"
	"os"
	"strings"
)

var CsvPath = "files/parsed/ifm.csv"
var PagesPath = "files/pages/ifm/"

func init() {

	CreateDir(PagesPath)
	err := CreateDir(PagesPath)
	if err != nil {
		log.Println(err)
	}

	err = CreateCsv(CsvPath)
	if err != nil {
		log.Println(err)
	}

}

func main() {
	Crawl()
}

func Crawl() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("sensoren.ru"),
		//colly.Async(),
	)

	// After making a request print "Visited ..."
	c.OnResponse(func(r *colly.Response) {

		page, err := OnPage(r)
		if err != nil {
			fmt.Println(err)
			return
		}

		if !page {
			return
		}
		fmt.Println("///////")
		fmt.Println("THIS IS THE PAGE!!!!", r.Request.URL)
		fmt.Println("///////")
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// convert relative url to absolute
		url := e.Request.AbsoluteURL(link)

		visit, err := OnLink(e)
		if err != nil {
			fmt.Println(err)
			return
		}

		if !visit {
			return
		}

		// Visit link found on page on a new thread
		c.Visit(url)
	})

	c.Visit("https://sensoren.ru/brands/ifm_electronic/")
}

func OnLink(e *colly.HTMLElement) (bool, error) {
	link := e.Attr("href")

	// convert relative url to absolute
	url := e.Request.AbsoluteURL(link)

	if strings.Contains(url, "brand-is-ifm_electronic") {
		return true, nil
	}

	if strings.Contains(url, "_ifm_electronic_") {
		return true, nil
	}

	if strings.Contains(url, "ifm_electronic") {
		return true, nil
	}

	return false, nil
}

func OnPage(e *colly.Response) (bool, error) {

	url := e.Request.URL.String()

	// if HTML contains div with class "product-page" then this is the page
	if !bytes.Contains(e.Body, []byte(`<div class="product-page">`)) {
		// just skip this url, no errors triggered
		return false, nil
	}

	h := Hash(url) + "html"

	err := WriteCsvP(CsvPath, []string{url, h})
	if err != nil {
		return true, err
	}

	err = SavePage(PagesPath+h, string(e.Body))
	if err != nil {
		return true, err
	}

	return true, nil
}

func SavePage(fstr string, html string) error {

	file, err := os.Create(fstr)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		return err
	}

	return nil
}

func WriteCsvP(path string, data []string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(data)

	return nil
}

func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateCsv(path string) error {

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	return nil
}

func Hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
