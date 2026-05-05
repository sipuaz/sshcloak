package config

const (
	EMPTY_STR                      = ""
	FILE_READ_WRITE_OWNER_READ_ALL = 0644
	COMMENT_CHAR                   = '#'
	EQUAL_CHAR                     = '='
	END_TAB_STR                    = " \t"
	BLANK_STR                      = " "
	NEGATION_STR                   = "!"
    ALL_STR                        = "All"
	ALL_LOWER_STR                  = "all"
	HOST_STR                       = "host"
	USER_STR                       = "user"
	LOCALUSER_STR                  = "localuser"
	STAR_STR                       = "*"
	MATCH_STR                      = "match"
	DOUBLE_QUOTE_CHAR              = '"'
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
