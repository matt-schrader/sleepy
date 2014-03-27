package sleepy

import (
    "fmt"
    "regexp"
)

func pathToRegexpString(routePath string) string {
  var re *regexp.Regexp
  regexpString := routePath

  // Dots
  re = regexp.MustCompile(`([^\\])\.`)
  regexpString = re.ReplaceAllStringFunc(regexpString, func(m string) string {
    return fmt.Sprintf(`%s\.`, string(m[0]))
  })

  // Wildcard names
  re = regexp.MustCompile(`:[^/#?()\.\\]+\*`)
  regexpString = re.ReplaceAllStringFunc(regexpString, func(m string) string {
    return fmt.Sprintf("(?P<%s>.+)", m[1:len(m) - 1])
  })

  re = regexp.MustCompile(`:[^/#?()\.\\]+`)
  regexpString = re.ReplaceAllStringFunc(regexpString, func(m string) string {
    return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:len(m)])
  })

  return fmt.Sprintf(`\A%s\z`, regexpString)
}

