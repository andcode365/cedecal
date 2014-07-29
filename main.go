package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	http.Handle("/cedec2014.ics", http.HandlerFunc(responser))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Print("ListenAndServe:", err)
	}
}

type speaker struct {
	Company_en string
	Company    string
	Name_en    string
	Name       string
}

type post struct {
	Speakers          []speaker
	Room              string
	Category_id       string
	Format_id         string
	Type_id           string
	Title             string
	Quick_description string
	Takeaway          string
	Held_at           string
}

type root struct {
	Posts []post
}

func responser(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(
		"https://cedec.cesa.or.jp/2014/session/data/json/download")
	if err != nil {
		log.Print("responser:", err)
	}

	defer resp.Body.Close()

	w.Header().Add("Content-Type", "text/calendar")

	dec := json.NewDecoder(resp.Body)
	var rt root
	dec.Decode(&rt)
	var buf bytes.Buffer
	convert(&rt, &buf)
	w.Write(buf.Bytes())
}

func arrange(desc string) (string, error) {
	dos := strings.Replace(desc, "\n", "", -1) // newline causes some troubles.

	var result string
	sum := 0
	for len(dos) > 0 {
		r, size := utf8.DecodeRuneInString(dos)
		if sum + size < 75 {
			result += string(r)
			sum += size
		} else {
			result += "\n " + string(r)
			sum = 1 + size
		}
		dos = dos[size:]
	}

	return result, nil
}

func writePost(p *post, buf *bytes.Buffer) {
	t, err := time.Parse("2006/01/02 15:04:05", p.Held_at)
	if err != nil {
		log.Println("err=", err)
		return
	}

	buf.WriteString("BEGIN:VEVENT\r\n")

	desc, err := arrange(p.Quick_description)
	title, err := arrange(p.Title)

	buf.WriteString("DESCRIPTION:" + desc + "\r\n")
	buf.WriteString("SUMMARY:" + title + "\r\n")
	buf.WriteString("DTSTART:" + t.Format("20060102T150405") + "\r\n")
	buf.WriteString("DTEND:" + t.Add(time.Hour).Format("20060102T150405") + "\r\n")

	buf.WriteString("END:VEVENT\r\n")
}

func convert(rt *root, buf *bytes.Buffer) {
	buf.WriteString("BEGIN:VCALENDAR\r\n")

	buf.WriteString("PRODID:\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")

	buf.WriteString("BEGIN:VTIMEZONE\r\n")
	{
		buf.WriteString("TZID:Japan\r\n")
		buf.WriteString("BEGIN:STANDARD\r\n")
		{
			buf.WriteString("DTSTART:19390101T000000\r\n")
			buf.WriteString("TZOFFSETFROM:+0900\r\n")
			buf.WriteString("TZOFFSETTO:+0900\r\n")
			buf.WriteString("TZNAME:JST\r\n")
		}
		buf.WriteString("END:STANDARD\r\n")
	}
	buf.WriteString("END:VTIMEZONE\r\n")

	for i := range rt.Posts {
		writePost(&rt.Posts[i], buf)
	}

	buf.WriteString("END:VCALENDAR\r\n")
}
