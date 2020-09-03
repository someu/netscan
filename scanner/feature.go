package scanner

import (
	"fmt"
	"regexp"
)

type FeatureVersion struct {
	Value string `json:"value"`
	If    string `json:"if"`
	Then  string `json:"then"`
	Else  string `json:"else"`
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

func (ruleItem *FeatureRuleItem) MatchContent(content string) (bool, string) {
	if len(content) == 0 {
		return false, ""
	}
	var matched bool
	var version string
	if len(ruleItem.Regexp) > 0 {
		re, err := regexp.Compile(fmt.Sprintf("(?i)%s", ruleItem.Regexp))
		if err != nil {
			fmt.Println(err, ruleItem.Regexp)
		} else {
			matched = re.MatchString(content)
			//if matched {
			//	fmt.Println("proof", ruleItem.Regexp, strings.Join(re.FindAllString(content, -1), ""))
			//}
		}
	}
	return matched, version
}

func (rule *FeatureRule) MatchResponse(response *Response) (bool, []string) {
	var matched = false
	var versions []string
	for _, ruleItem := range rule.Title {
		currentMatched, version := ruleItem.MatchContent(response.Title)
		matched = matched || currentMatched
		if len(version) > 0 {
			versions = append(versions, version)
		}
	}

	for _, ruleItem := range rule.Header {
		currentMatched, version := ruleItem.MatchContent(response.Header)
		matched = matched || currentMatched
		if len(version) > 0 {
			versions = append(versions, version)
		}
	}

	for _, ruleItem := range rule.Cookie {
		currentMatched, version := ruleItem.MatchContent(response.Cookie)
		matched = matched || currentMatched
		if len(version) > 0 {
			versions = append(versions, version)
		}
	}

	for _, ruleItem := range rule.Body {
		currentMatched, version := ruleItem.MatchContent(response.Body)
		matched = matched || currentMatched
		if len(version) > 0 {
			versions = append(versions, version)
		}
	}

	for key, ruleItems := range rule.MetaTag {
		for _, ruleItem := range ruleItems {
			currentMatched, version := ruleItem.MatchContent(response.MetaTag[key])
			matched = matched || currentMatched
			if len(version) > 0 {
				versions = append(versions, version)
			}
		}
	}

	for key, ruleItems := range rule.CookieField {
		for _, ruleItem := range ruleItems {
			currentMatched, version := ruleItem.MatchContent(response.CookieField[key])
			matched = matched || currentMatched
			if len(version) > 0 {
				versions = append(versions, version)
			}
		}
	}

	for key, ruleItems := range rule.HeaderField {
		for _, ruleItem := range ruleItems {
			currentMatched, version := ruleItem.MatchContent(response.HeaderField[key])
			matched = matched || currentMatched
			if len(version) > 0 {
				versions = append(versions, version)
			}
		}
	}

	return matched, versions
}
