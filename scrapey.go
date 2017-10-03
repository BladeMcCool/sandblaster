package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/kr/pretty"
	_ "github.com/nakagami/firebirdsql"
	"github.com/spaolacci/murmur3"
	_ "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cacheDir string = "cache"

const maxCacheAge int = 1200

// const maxCacheAge int = 300

func scrapey() {
	fmt.Printf("Yeah herely.\n")
	// startUrl := "https://example.com"
	// doc, err := goquery.NewDocument(startUrl)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(doc.HTML())

	// doc.Find("a.menu").Each(func(i int, item *goquery.Selection) {
	// 	fmt.Println(i, item.Text())
	// })

	hinkyProxies := hinkyDink()
	_ = hinkyProxies
	proxyDb := connectDb()
	defer proxyDb.Close()
	saveProxies(proxyDb, hinkyProxies)
	printAllproxies(proxyDb)

	//read the first page, url set here.
	//check our local cache first, if page is not in it or is old, get fresh.
	//get a list of pages to get. or start at the first one and keep going until they are all gotten.
	//on each page, extract the proxy info
	//save the proxy info somewhere. for now, a text file is fine.

}

func saveProxies(conn *sql.DB, proxies []proxyInfo) {
	for _, proxy := range proxies {
		_, err := conn.Exec("INSERT INTO proxies (ip, port, country, foundon) VALUES (?,?,?,?)", proxy.ip, proxy.port, proxy.claimedCountry, proxy.foundOn)
		if err != nil {
			log.Printf("sql err? %s", err.Error())
		}
	}
}
func printAllproxies(conn *sql.DB) {
	rows, err := conn.Query("SELECT ip, port, country, foundon FROM proxies ORDER BY country, ip")
	defer rows.Close()
	if err != nil {
		log.Printf("sql err? %s", err.Error())
	}
	for rows.Next() {
		proxy := proxyInfo{}
		err := rows.Scan(&proxy.ip, &proxy.port, &proxy.claimedCountry, &proxy.foundOn)
		if err != nil {
			log.Printf("sql err %s", err.Error())
		}
		log.Printf("row: %#v", proxy)
	}
	log.Printf("results: ")
}

func connectDb() *sql.DB {
	// var n int
	// conn, err := sql.Open("firebirdsql", `SYSDBA:asshat1234@localhost/C:\Program\ Files\Firebird\Firebird_3_0\PROXIES.FDB`)
	// conn, connErr := sql.Open("firebirdsql", `SYSDBA:asshat1234@localhost/c:\\Program Files\\Firebird\\Firebird_3_0\\proxies.fdb`)
	// conn, connErr := sql.Open("firebirdsql", `SYSDBA:asshat1234@localhost/c:\\Program Files\\Firebird\\Firebird_3_0\\proxies.fdb`)
	conn, connErr := sql.Open("firebirdsql", `sysdba:asshat1234@localhost:3050/c:/fbdata/proxies.fdb`)
	// defer conn.Close()
	if connErr != nil {
		log.Fatalf("error from sql: %s", connErr.Error())
		log.Println("errro wot")
	}
	log.Println("guess no error!!")

	// conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
	// goddammit := conn.QueryRow("SELECT Count(*) FROM proxies").Scan(&n)
	// if goddammit != nil {
	// 	log.Fatalf("error2 from sql: %s", goddammit.Error())
	// 	log.Println("errro2 wot")
	// }
	// // conn.QueryRow("SELECT Count(*) FROM rdb$relations")
	// fmt.Println("Proxies record count=", n)
	// _ = n
	return conn
}
func hinkyDink() []proxyInfo {
	hinkyBase := "http://www.mrhinkydink.com/"
	startUrl := hinkyBase + "proxies.htm"
	firstPage := getDoc(startUrl)
	if firstPage == nil {
		log.Fatalf("no first page of hinky dink. fatality\n")
	}

	proxyPageRe := regexp.MustCompile(`Proxies page (\d+)`)
	// if err != nil {
	// 	log.Fatalf("regexp err: %s\n", err.Error())
	// }
	resultPages := []*goquery.Document{
		firstPage,
	}

	log.Printf("here1\n")
	firstPage.Find("a.menu").Each(func(i int, item *goquery.Selection) {
		log.Printf("found a menu item ...\n")
		linkPath, _ := item.Attr("href")
		linkText := item.Text()
		matchInfo := proxyPageRe.FindStringSubmatch(linkText)
		log.Printf("extract of matching menuitem text: % #v\n", pretty.Formatter(matchInfo))
		if len(matchInfo) != 2 {
			return
		}
		fullLink := hinkyBase + linkPath
		pageNum, _ := strconv.Atoi(string(matchInfo[1]))
		if pageNum <= 1 {
			return
		}
		log.Printf("url for page %d is %s (will sleep and then fetch its contents.)\n", pageNum, fullLink)
		time.Sleep(1 * time.Second)
		pageDocument := getDoc(fullLink)
		resultPages = append(resultPages, pageDocument)
		// for ind, fucker := range matchInfo {
		// 	fmt.Printf("something: %d %s\n", ind, fucker)
		// }
		// // fmt.Printf("something: % #v\n", pretty.Formatter(something))
		// fmt.Println(i, linkText, linkPath, )
	})
	log.Printf("here3, fetched %d pages of proxy data to go through.\n", len(resultPages))

	hinkyAllProxies := []proxyInfo{}
	for i, page := range resultPages {
		_ = page
		log.Printf("going through page %d ...\n", i)
		hinkyAllProxies = append(hinkyAllProxies, hinkyDinkPageExtract(page)...)
	}
	log.Printf("got a number %d of this: % #v\n", len(hinkyAllProxies), pretty.Formatter(hinkyAllProxies))
	return hinkyAllProxies
}

type proxyInfo struct {
	ip             string
	port           int
	claimedCountry string
	foundOn        string
}

func hinkyDinkPageExtract(doc *goquery.Document) []proxyInfo {
	pageProxies := []proxyInfo{}
	doc.Find("tr.text").Each(func(i int, tr *goquery.Selection) {
		log.Printf("This tr has %d nodes.\n", len(tr.Nodes))
		foundTds := tr.Find("td")
		log.Printf("This trs foundTds has %d nodes.\n", len(foundTds.Nodes))
		if foundTds.Length() != 8 {
			return
		}

		// debug
		// foundTds.Each(func(j int, td *goquery.Selection) {
		// 	html, _ := td.Html()
		// 	text := td.Text()
		// 	_ = html
		// 	log.Printf("something td: %d '%s' and '%s' \n", j, html, text)
		// })
		//end debug

		port, _ := strconv.Atoi(foundTds.Eq(1).Text())
		proxy := proxyInfo{
			ip:             foundTds.Eq(0).Text(),
			port:           port,
			claimedCountry: strings.TrimSpace(foundTds.Eq(4).Text()),
			foundOn:        "hinkyDink",
		}

		pageProxies = append(pageProxies, proxy)
	})
	return pageProxies
}

func getDoc(url string) *goquery.Document {
	localCacheFileName := urlToFilename(url)
	isCached := checkLocalCache(localCacheFileName)

	if isCached {
		log.Printf("Should read from the local cached copy.\n")
		return docFromCache(localCacheFileName)
	}
	doc := docFromUrl(url)
	cacheDoc(doc, url)
	return doc
}

func urlToFilename(url string) string {
	//maybe we should hash it.
	//for now lets log the filename conversion too so we can tell what the hashes were
	localFilename := fmt.Sprintf("%x.cached", murmur3.Sum64([]byte(url)))
	fmt.Printf("DBG url hash %s from %s\n", localFilename, url)
	return localFilename
}

func checkLocalCache(filename string) bool {
	//if filename exists in local cache dir, then return a io reader to read the file contents.
	cachePath := filepath.Join(".", cacheDir)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Printf("Created cache directory at %s\n", cachePath)
		os.Mkdir(cachePath, os.ModePerm)
	}

	//file not exist?
	filePath := filepath.Join(cachePath, filename)
	log.Printf("Checking if cache file %s exists\n", filePath)
	fileStats, err := os.Stat(filePath)
	// var err error
	if err != nil && os.IsNotExist(err) {
		log.Printf("file %s does not exist\n", filePath)
		return false
	}

	// log.Printf("File stats: %# v\n", pretty.Formatter(fileStats))
	fileAgeSec := time.Since(fileStats.ModTime()).Seconds()
	log.Printf("File age sec %f\n", fileAgeSec)
	if int(fileAgeSec) > maxCacheAge {
		log.Printf("File age sec %f is too old\n", fileAgeSec)

		log.Printf("DBG BUT FOR NOW LETS JUST USE IT")
		// return false

	}

	return true
}

func docFromCache(filename string) *goquery.Document {
	fpath := filepath.Join(".", cacheDir, filename)
	log.Printf("docFromCache got io.reader for %s\n", fpath)
	reader, _ := os.Open(fpath)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
func docFromUrl(url string) *goquery.Document {
	log.Printf("Loading document from url %s\n", url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
func cacheDoc(doc *goquery.Document, url string) {
	fn := urlToFilename(url)
	fpath := filepath.Join(".", cacheDir, fn)
	html, _ := doc.Html()
	err := ioutil.WriteFile(fpath, []byte(html), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to write file to cache at %s.\n", fpath)
	}
	log.Printf("Cached url %s to local cache file %s\n", url, fn)
}
