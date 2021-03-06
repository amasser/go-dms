package controllers

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	constant "github.com/ninadakolekar/go-dms/src/constants"
	solr "github.com/rtt/Go-Solr"
)

type lINK struct {
	DocName   string
	Idate     string
	DocId     string
	CreateTS  string
	ApproveTS string
	DocType   string
	EffDate   string
	ExpDate   string
}
type typeSort []lINK
type expDateSort []lINK
type effDateSort []lINK
type docIdSort []lINK
type docNameSorter []lINK
type idateSorter []lINK
type appTSsort []lINK
type createTSsort []lINK

func (a typeSort) Len() int           { return len(a) }
func (a typeSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a typeSort) Less(i, j int) bool { return a[i].DocType < a[j].DocType }

func (a expDateSort) Len() int           { return len(a) }
func (a expDateSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a expDateSort) Less(i, j int) bool { return a[i].ExpDate < a[j].ExpDate }

func (a effDateSort) Len() int           { return len(a) }
func (a effDateSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a effDateSort) Less(i, j int) bool { return a[i].EffDate < a[j].EffDate }

func (a docIdSort) Len() int           { return len(a) }
func (a docIdSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a docIdSort) Less(i, j int) bool { return a[i].DocId < a[j].DocId }

func (a docNameSorter) Len() int           { return len(a) }
func (a docNameSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a docNameSorter) Less(i, j int) bool { return a[i].DocName < a[j].DocName }

func (a idateSorter) Len() int           { return len(a) }
func (a idateSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a idateSorter) Less(i, j int) bool { return a[i].Idate < a[j].Idate }

func (a appTSsort) Len() int           { return len(a) }
func (a appTSsort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a appTSsort) Less(i, j int) bool { return a[i].ApproveTS > a[j].ApproveTS }

func (a createTSsort) Len() int           { return len(a) }
func (a createTSsort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a createTSsort) Less(i, j int) bool { return a[i].CreateTS > a[j].CreateTS }

//ProcessDocASearch ... process doc search
func ProcessDocASearch(w http.ResponseWriter, r *http.Request) {

	alert := false
	data := false
	alertmsg := "no msg"
	links := []lINK{}
	if r.Method == "POST" {
		r.ParseForm()
		sCriteria := []string{"*", "*", "*", "*", "*", "*", "*"}
		sKeyword := []string{"*", "*", "*", "*", "*", "*", "*"}
		sortOrder := html.EscapeString(r.Form["sort"][0])

		for i := 0; i < 6; i++ {
			if len(r.Form["criteria"+strconv.Itoa(i+1)]) > 0 {
				sCriteria[i] = html.EscapeString(r.Form["criteria"+strconv.Itoa(i+1)][0])
				sKeyword[i] = html.EscapeString(r.Form["searchKeyword"+strconv.Itoa(i+1)][0])
			}
		}
		for i := 0; i < 6; i++ {
			sKeyword[i] = removeIntialEndingspaces(sKeyword[i])
		}

		if validateSearchForm(sCriteria, sKeyword) == true {

			query := makeSearchQuery(sCriteria, sKeyword)

			s, err := solr.Init(constant.SolrHost, constant.SolrPort, constant.DocsCore)
			if err != nil {
				fmt.Println(err)
			}

			//	fmt.Println(query) //Debug
			q := solr.Query{
				Params: solr.URLParamMap{
					"q": []string{query},
				},
				Rows: 100,
			}

			res, err := s.Select(&q)
			if err != nil {
				fmt.Println(err)

			}

			results := res.Results
			if results.Len() == 0 {
				alert = true
				alertmsg = "No Results Found!"
			} else {
				for i := 0; i < results.Len(); i++ {
					links = append(links, convertTolINK(results.Get(i)))
				}
				data = true

				links = sortby(links, sortOrder)
				fmt.Println("after sorting \n", links) //Debug
			}

		} else {
			alert = true
			alertmsg = "Invalid Search Query!"
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/searchDoc.html"))
	tmpl.Execute(w, struct {
		Alertb   bool
		Alertmsg string
		Datab    bool
		Data     []lINK
	}{alert, alertmsg, data, links})
}

func sortby(l []lINK, so string) []lINK {
	//a => acsending d => decending
	if so == "alexical" {
		sort.Sort(docNameSorter(l))
	} else if so == "aTime" {
		sort.Sort(idateSorter(l))
	} else if so == "dApprovedTS" {
		sort.Sort(appTSsort(l))
	} else if so == "dCreateTS" {
		sort.Sort(createTSsort(l))
	} else if so == "alexicalId" {
		sort.Sort(docIdSort(l))
	} else if so == "typeSort" {
		sort.Sort(typeSort(l))
	} else if so == "effDate" {
		sort.Sort(effDateSort(l))
	} else if so == "expDate" {
		sort.Sort(expDateSort(l))
	}
	return l
}

func makeSearchQuery(sC []string, sK []string) string {
	validCriterion := []string{"docNumber", "docName", "docKeyword", "initiator", "creator", "reviewer", "approver", "auth", "dept", "from Init Date", "from Eff Date", "from Exp Date", "till Init Date", "till Eff Date", "till Exp Date"}
	validQueryPrifex := []string{"id:", "title:", "body:", "initiator:", "creator:", "reviewer:", "approver:", "authorizer:", "docDepartment:", "initTime:", "effDate:", "expDate:", "initTime:", "effDate:", "expDate:"}

	querys := []string{}
	counter := 0
	for i, sc := range sC {
		for j, v := range validCriterion {
			if v == sc {

				if j == 10 || j == 9 || j == 11 {
					querys = append(querys, validQueryPrifex[j]+"["+sK[i]+"T23:59:59Z TO *]")
				} else if j == 12 || j == 13 || j == 14 {
					querys = append(querys, validQueryPrifex[j]+"[ 2000-01-01T00:00:58Z TO "+sK[i]+"T23:59:59Z ]")
				} else if j == 0 || j == 1 {
					querys = append(querys, validQueryPrifex[j]+sK[i]+"*")
				} else if j == 2 {
					strs := strings.Split(sK[i], " ")
					str := "*"
					for _, e := range strs {
						str = str + e + "*"
					}
					querys = append(querys, validQueryPrifex[j]+str)
				} else {
					querys = append(querys, validQueryPrifex[j]+sK[i])
				}
				counter++
			}
		}
	}
	query := ""
	for i, q := range querys {
		if i == 0 {
			query = q
		} else {
			query += (" AND " + q)
		}
	}

	return query
}
func isDatevalid(s string) bool {
	y, err := strconv.Atoi(s[0:4])
	if err != nil {
		return false
	}

	m, err := strconv.Atoi(s[5:7])
	if err != nil {
		return false
	}
	d, err := strconv.Atoi(s[8:10])
	if err != nil {
		return false
	}

	if (y%4 == 0 && y%100 == 0) || (y%4 != 0) {
		if m == 1 || m == 3 || m == 5 || m == 7 || m == 8 || m == 10 || m == 12 {
			if d < 0 || d > 31 {
				return false
			}
		} else if m == 4 || m == 6 || m == 9 || m == 11 {
			if d < 0 || d > 30 {
				return false
			}
		} else if m == 2 {
			if d < 0 || d > 29 {
				return false
			}
		} else {
			return false
		}
	} else {
		if m == 1 || m == 3 || m == 5 || m == 7 || m == 8 || m == 10 || m == 12 {
			if d < 0 || d > 31 {
				return false
			}
		} else if m == 4 || m == 6 || m == 9 || m == 11 {
			if d < 0 || d > 30 {
				return false
			}
		} else if m == 2 {
			if d < 0 || d > 28 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func validateSearchForm(sC []string, sK []string) bool {
	validCriterion := []string{"docNumber", "docName", "docKeyword", "initiator", "creator", "reviewer", "approver", "auth", "dept", "from Init Date", "from Eff Date", "from Exp Date", "till Init Date", "till Eff Date", "till Exp Date"}
	isKeyword := regexp.MustCompile(`^[A-Za-z0-9 ]+$`).MatchString
	isDate := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString
	isAlphaNumeric := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString

	for j, sc := range sC {

		for i, v := range validCriterion {

			if sc == v {

				if i == 9 || i == 10 || i == 11 || i == 12 || i == 13 || i == 14 {
					if isDatevalid(sK[j]) == false || !isDate(sK[j]) {
						return false
					}
				} else if i == 2 {
					if isKeyword(sK[j]) == false {
						return false
					}
				} else {
					if isAlphaNumeric(sK[j]) == false {
						return false
					}
				}

			}
		}
	}
	return true
}

func removeIntialEndingspaces(str string) string {
	s := 0
	e := len(str) - 1

	for ; str[s] == ' '; s++ {

	}
	for e = len(str) - 1; str[e] == ' '; e-- {

	}

	return str[s : e+1]
}

func convertTolINK(s1 *solr.Document) lINK {
	ctime := ""
	atime := ""
	createTime := s1.Field("createTime")
	if createTime != nil {
		ctime = s1.Field("createTime").(string)
	}
	approvetime := s1.Field("approveTime")
	if approvetime != nil {
		atime = s1.Field("approvetime").(string)
	}
	return lINK{s1.Field("title").(string), s1.Field("initTime").(string), s1.Field("id").(string), ctime, atime, s1.Field("docType").(string), s1.Field("effDate").(string), s1.Field("expDate").(string)}
}
