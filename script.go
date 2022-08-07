package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Persoon struct {
	Context string `json:"@odata.context"`
	Id string
	Nummer int
	Titels string
	Initialen string
	Tussenvoegsel string
	Achternaam string
	Voornamen string
	Roepnaam string
	Geslacht string
	Functie string
	Geboortedatum string
	Geboorteplaats string
	Geboorteland string
	Overlijdensdatum string
	Overlijdensplaats string
	Woonplaats string
	Land string
	Fractielabel string
	ContentType string
	ContentLength int
	GewijzigdOp string
	ApiGewijzigdOp string
	Verwijderd bool
}

type PersoonGeschenk struct {
	Id string
	Omschrijving string
	Datum string
	Gewicht int
	GewijzigdOp string
	ApiGewijzigdOp string
	Verwijderd bool
	Persoon_Id string
}

type PersoonGeschenkResponse struct {
	Context string `json:"@odata.context"`
	Value []PersoonGeschenk
	NextLink string `json:"@odata.nextLink"`
}

func get_geschenken_count() {
	response, err := http.Get("https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0/PersoonGeschenk?$count=true")
	if err != nil {
	   log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
  	 log.Fatalln(err)
	}

	fmt.Println("Count is: ", string(body))
}

func get_geschenken(nextLink string) PersoonGeschenkResponse {
	var response *http.Response
	var err error

	if (nextLink != "first") {
		response, err = http.Get(nextLink)
	} else {
		response, err = http.Get("https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0/PersoonGeschenk")
	}
	if err != nil {
	   log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
  	 log.Fatalln(err)
	}

	var r PersoonGeschenkResponse
	json.Unmarshal(body, &r)
	return r
}

func get_persoon(uid string) Persoon {
	response, err := http.Get("https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0/Persoon/"+uid)
	if err != nil {
	   log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var p Persoon
	json.Unmarshal(body, &p)
	return p
}

func is_price_character(c byte, dot bool) bool {
	if ((c >= '0' && c <= '9')) {
		return true
	}
	if (!dot && (c == '.' || c == ' ')) {
		return true
	}
	return false
}

func strip_non_price_characters(s string) string {
	lastIndex := 0
	haveFoundDot := false

	s = strings.ReplaceAll(s,",", ".")

	for i := 0; is_price_character(s[i], haveFoundDot) && i < len(s) - 1; i++ {
		if s[i] == '.' {
			haveFoundDot = true
		}
		lastIndex = i
	}
	return s[:lastIndex]
}

func parse_waarde(omschrijving string) float64 {
	s := strings.Split(omschrijving, "€")
	if (len(s) == 1) {
		return 0
	}
	val := strip_non_price_characters(s[1])
	res, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if (err != nil) {
		fmt.Println(err.Error())
		return 0
	}
	return res
}

func sort_totals(totals map[string]float64) []string {
	keys := make([]string, 0, len(totals))
  
    for key := range totals {
        keys = append(keys, key)
    }
  
    sort.SliceStable(keys, func(i, j int) bool{
        return totals[keys[i]] > totals[keys[j]]
    })
  
    return keys
}

func main() {
	// get_geschenken_count()
	totals := make(map[string]float64)

	var geschenken PersoonGeschenkResponse
	geschenken.NextLink = "first"
	for geschenken.NextLink != "" {
		geschenken = get_geschenken(geschenken.NextLink)
		for _, g := range geschenken.Value {
			waarde := parse_waarde(g.Omschrijving)
			totals[g.Persoon_Id] += waarde
		}
	}

	keys := sort_totals(totals)
	for _, k := range keys {
		p := get_persoon(k)
		fmt.Println(p.Roepnaam, p.Tussenvoegsel, p.Achternaam)
		fmt.Println("€", math.Floor(totals[k]*100)/100)
		fmt.Println()
	}
}