package common

import (
	"os"
)

// DATABASE GLOBALS
const TRANSP_AEROL = 0
const TRANSP_LAN = 1
const TRANSP_BUS = 2

// TRANSFER TIME (mins)
const TRANSFER_TIME = 120

// MAX HOURS OF TRANSFER
const MAX_TRANSFER_HOURS = 12

// MAX TRANSFERS
const MAX_TRANSFERS = 5

// NUMBER OF DAYS AWAY FROM DEP DAY FOR SPECIFIC TRIP LOOKUP
const MAX_DAYS_SPEC = 4

// // PATH TO HTML FILES
const WEB_HTML_PATH = `resources` + string(os.PathSeparator) + `html` + string(os.PathSeparator)

// PATH TO WEB RESOURCE ELEMENTS
const WEB_STATIC_PATH = `resources` + string(os.PathSeparator) + `web_resources` + string(os.PathSeparator)

const HEURISTIC_SPEC_PATH = `resources` + string(os.PathSeparator) + `static` + string(os.PathSeparator) + `heuristicSpec.gob`
const HEURISTIC_GEN_PATH = `resources` + string(os.PathSeparator) + `static` + string(os.PathSeparator) + `heuristicGen.gob`

// NUMBER OF DAYS AWAY FROM THE DEPARTURE DAY TO LOOK FOR IN GENERAL PATHS
const GEN_DAYS_SCOPE = 1

const FILE_SPEC_HEURISTIC = "heuristicSpec.gob"
const FILE_GEN_HEURISTIC = "heuristicGen.gob"
