package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"phonebook/lib"
	"rentroll/rlib"
	"strings"
	"time"
)
import _ "github.com/go-sql-driver/mysql"

// CmdRUNBOOKS and the rest are command numbers used by the Dispatch function.
const (
	CmdRUNBOOKS     = 1 // Run Journal and ledgers over a defined period
	CmdTRIALBALANCE = 5 // balance of all ledgers at the end of the defined period

	FMTTEXT = 1 // output format is text
	FMTHTML = 2 // output format is html
)

// DispatchCtx is a type of struct needed for the Dispatch function. It defines
// everything needed to run a particular command. It is the responsibility of the
// caller to fill out all the needed ctx information. Not all information is needed
// for all commands.
type DispatchCtx struct {
	Cmd          int                 // cmd to execute
	DtStart      time.Time           // period start time
	DtStop       time.Time           // period end
	OutputFormat int                 // how shall the output be formatted
	Report       int64               // which report to generate - this is used in batch mode operation
	w            http.ResponseWriter // for web responses
	r            *http.Request       // for web request introspection
}

// App is the global data structure for this app
var App struct {
	dbdir     *sql.DB  // phonebook db
	dbrr      *sql.DB  // rentroll db
	DBDir     string   // phonebook database
	DBRR      string   // rentroll database
	PortRR    int      // port on which rentroll listens
	DBUser    string   // user for all databases
	Report    int64    // if testing engine, which report/action to perform
	bizfile   string   // TEMPORARY - tests loading bizcsv
	LogFile   *os.File // where to log messages
	BatchMode bool     // if true, then don't start http, the command line request is for a batch process
	sStart    string   //start time
	sStop     string   //stop time
	// AsmtTypes map[int64]rlib.AssessmentType
	// PmtTypes  map[int64]rlib.PaymentType
}

// RRfuncMap is a map of functions passed to each html page that can be referenced
// as needed to produce the page
var RRfuncMap map[string]interface{}

// Chttp is a server mux for handling unprocessed html page requests.
// For example, a .css file or an image file.
var Chttp = http.NewServeMux()

func readCommandLineArgs() {
	dbuPtr := flag.String("B", "ec2-user", "database user name")
	dbnmPtr := flag.String("N", "accord", "directory database (accord)")
	dbrrPtr := flag.String("M", "rentroll", "database name (rentroll)")
	pStart := flag.String("j", "2015-11-01", "Accounting Period start time")
	pStop := flag.String("k", "2015-12-01", "Accounting Period end time")
	verPtr := flag.Bool("v", false, "prints the version to stdout")
	bizPtr := flag.String("b", "b.csv", "add Business via csv file")
	rptPtr := flag.Int64("r", 0, "report: 0 = generate Journal records, 1 = Journal, 2 = Rentable")
	portPtr := flag.Int("p", 8270, "port on which RentRoll server listens")
	bPtr := flag.Bool("A", false, "if specified run as a batch process, do not start http")

	flag.Parse()
	if *verPtr {
		fmt.Printf("Version:    %s\nBuild Time: %s\n", getVersionNo(), getBuildTime())
		os.Exit(0)
	}
	App.DBDir = *dbnmPtr
	App.DBRR = *dbrrPtr
	App.DBUser = *dbuPtr
	App.Report = *rptPtr
	App.bizfile = *bizPtr
	App.sStart = *pStart
	App.sStop = *pStop
	App.PortRR = *portPtr
	App.BatchMode = *bPtr
}

func intTest(xbiz *rlib.XBusiness, d1, d2 *time.Time) {
	fmt.Printf("INTERNAL TEST\n")
	m := rlib.ParseAcctRule(xbiz, 1, d1, d2, "d ${DFLTGENRCV} 1000.0, c 40001 ${UMR}, d 41004 ${UMR} ${aval(${DFLTGENRCV})} -", float64(1000), float64(8)/float64(30))

	for i := 0; i < len(m); i++ {
		fmt.Printf("m[%d] = %#v\n", i, m[i])
	}
	fmt.Printf("DONE\n")
}

// HomeHandler serves static http content such as the .css files
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".") {
		Chttp.ServeHTTP(w, r)
	} else {
		http.Redirect(w, r, "/dispatch/", http.StatusFound)
	}
}

func initHTTP() {
	Chttp.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/dispatch/", dispatchHandler)
	http.HandleFunc("/srvformTrialBal/", srvformTrialBalance)
	http.HandleFunc("/trialbalance/", hndTrialBalance)
}

func main() {
	readCommandLineArgs()
	rlib.RRReadConfig()

	//==============================================
	// Open the logfile and begin logging...
	//==============================================
	var err error
	App.LogFile, err = os.OpenFile("rentroll.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	lib.Errcheck(err)
	defer App.LogFile.Close()
	log.SetOutput(App.LogFile)
	rlib.Ulog("*** Accord RENTROLL ***\n")

	//----------------------------
	// Open RentRoll database
	//----------------------------
	// s := fmt.Sprintf("%s:@/%s?charset=utf8&parseTime=True", DBUser, DBRR)
	s := rlib.RRGetSQLOpenString(App.DBRR)
	App.dbrr, err = sql.Open("mysql", s)
	if nil != err {
		fmt.Printf("sql.Open for database=%s, dbuser=%s: Error = %v\n", App.DBRR, rlib.AppConfig.RRDbuser, err)
		os.Exit(1)
	}
	defer App.dbrr.Close()
	err = App.dbrr.Ping()
	if nil != err {
		fmt.Printf("DBRR.Ping for database=%s, dbuser=%s: Error = %v\n", App.DBRR, rlib.AppConfig.RRDbuser, err)
		os.Exit(1)
	}

	//----------------------------
	// Open Phonebook database
	//----------------------------
	s = rlib.RRGetSQLOpenString(App.DBDir)
	App.dbdir, err = sql.Open("mysql", s)
	if nil != err {
		fmt.Printf("sql.Open: Error = %v\n", err)
		os.Exit(1)
	}
	err = App.dbdir.Ping()
	if nil != err {
		fmt.Printf("dbdir.Ping: Error = %v\n", err)
		os.Exit(1)
	}

	rlib.InitDBHelpers(App.dbrr, App.dbdir)
	initRentRoll()

	if App.BatchMode {
		ctx := createStartupCtx()
		Dispatch(&ctx)
	} else {
		initHTTP()
		rlib.Ulog("RentRoll initiating HTTP service on port %d\n", App.PortRR)
		err = http.ListenAndServe(fmt.Sprintf(":%d", App.PortRR), nil)
		if nil != err {
			fmt.Printf("*** Error on http.ListenAndServe: %v\n", err)
			rlib.Ulog("*** Error on http.ListenAndServe: %v\n", err)
			os.Exit(1)
		}
	}
}
