package config

// File and directory stuff
const (
	RWOwnerRAll  = 0644
	RWXOwnerRAll = 0755
	currentDir   = "."
)

// SSH config syntax stuff
const (
	commentChar          = '#'
	commentHeader        = string(commentChar) + blankStr
	equalChar            = '='
	negationStr          = "!"
	allStr               = "All"
	allLowerStr          = "all"
	hostStr              = "host"
	userStr              = "User"
	userLowerStr         = "user"
	localUserStr         = "localuser"
	starStr              = "*"
	matchStr             = "match"
	includeHeader        = "Include "
	matchHeader          = "Match "
	hostHeader           = "Host "
	jollyHostStr         = "*?"
	hostNameStr          = "HostName"
	hostNameLowerStr     = "hostname"
	portStr              = "Port"
	portLowerStr         = "port"
	identityFileStr      = "IdentityFile"
	identityFileLowerStr = "identityfile"
)

// Rendered config formatting stuff
const (
	emptyString         = ""
	endTabStr           = " \t"
	blankStr            = " "
	doubleQuoteChar     = '"'
	doubleQuoteStr      = string(doubleQuoteChar)
	newLineChar         = '\n'
	newLineStr          = string(newLineChar)
	quotableChars       = " \t\""
	backslashStr        = "\\"
	escapedBackslashStr = "\\\\"
	escapedQuoteStr     = "\\\""
)

var directives = map[string]bool{
	"host":           true,
	"match":          true,
	"localforward":   true,
	"remoteforward":  true,
	"dynamicforward": true,
	"identityfile":   true,
	"sendenv":        true,
	"setenv":         true,
}
