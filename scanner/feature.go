package scanner

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type FeatureVersion struct {
	Value string `json:"value"`
	Match string `json:"match"`
}

type FeatureRuleItem struct {
	Regexp  string          `json:"regexp"`
	Md5     string          `json:"md5"`
	Version *FeatureVersion `json:"version"`
}

type FeatureRule struct {
	Name        string                        `json:"name"`
	Title       []*FeatureRuleItem            `json:"title"`
	Header      []*FeatureRuleItem            `json:"header"`
	Cookie      []*FeatureRuleItem            `json:"cookie"`
	MetaTag     map[string][]*FeatureRuleItem `json:"metaTag"`
	HeaderField map[string][]*FeatureRuleItem `json:"headerField"`
	CookieField map[string][]*FeatureRuleItem `json:"cookieField"`
	Body        []*FeatureRuleItem
}

type Feature struct {
	Path  string         `json:"path"`
	Rules []*FeatureRule `json:"rules"`
}

func (ruleItem *FeatureRuleItem) MatchContent(content string) (bool, []string) {
	if len(content) == 0 {
		return false, nil
	}
	var matched bool
	var versions []string
	if len(ruleItem.Regexp) == 0 {
		return false, nil
	}
	re, err := regexp.Compile(fmt.Sprintf("(?i)%s", ruleItem.Regexp))
	if err != nil {
		return false, nil
	}
	matched = re.MatchString(content)
	//if matched {
	//	fmt.Println("proof", ruleItem.Regexp, strings.Join(re.FindAllString(content, -1), ""))
	//}
	if matched && ruleItem.Version != nil {
		if len(ruleItem.Version.Match) >= 0 {
			versionRe, err := regexp.Compile(ruleItem.Version.Match)
			if err == nil {
				versions = versionRe.FindAllString(content, -1)
			}
		} else if len(ruleItem.Version.Value) > 0 {
			sss := re.FindAllStringSubmatch(content, -1)
			for _, ss := range sss {
				if len(ss) >= 2 && len(ss[1]) > 0 {
					versions = append(versions, strings.Replace(ruleItem.Version.Value, "\\1", ss[1], -1))
				}
			}
		}
	}

	return matched, versions
}

func (rule *FeatureRule) MatchResponse(response *Response) *MatchedApp {
	var matched = false
	var versions []string
	responseType := reflect.TypeOf(*response)
	responseValue := reflect.ValueOf(*response)
	ruleValue := reflect.ValueOf(*rule)
	stringType := reflect.TypeOf("")

	for i := 0; i < responseType.NumField(); i++ {
		responseField := responseType.Field(i)
		responseValue := responseValue.Field(i)
		responseValueType := responseValue.Type()
		ruleValue := ruleValue.FieldByName(responseField.Name).Interface()

		if responseValueType == stringType {
			// handle values(Title, Header, Cookie, Body) in response which type is string
			for _, ruleItem := range ruleValue.([]*FeatureRuleItem) {
				currentMatched, currentVersions := ruleItem.MatchContent(responseValue.Interface().(string))
				matched = matched || currentMatched
				if len(currentVersions) > 0 {
					versions = append(versions, currentVersions...)
				}
			}
		} else if responseValueType == reflect.MapOf(stringType, stringType) {
			// handle values(MetaTag, HeaderField, CookieField) in response which type is map[string]string
			for key, ruleItems := range ruleValue.(map[string][]*FeatureRuleItem) {
				for _, ruleItem := range ruleItems {
					currentMatched, currentVersions := ruleItem.MatchContent(responseValue.Interface().(map[string]string)[key])
					matched = matched || currentMatched
					if len(currentVersions) > 0 {
						versions = append(versions, currentVersions...)
					}
				}
			}
		} else {
			log.Println("not recognized type:", responseValueType.String())
		}
	}

	if matched {
		return &MatchedApp{Name: rule.Name, Versions: versions}
	} else {
		return nil
	}
}
