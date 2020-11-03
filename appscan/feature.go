package appscan

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type FeatureRule struct {
	Regexp       string
	regexp       *regexp.Regexp
	Md5          string
	VersionStock string
	versionStock *regexp.Regexp
	Version      string
}

type ruleField struct {
	Name string
	Type string
}

type Feature struct {
	ID uint
	// 特征名
	Name             string
	Path             string
	From             []string
	Types            []string
	Implies          []string
	ManufacturerName []string
	ManufacturerUrl  []string
	// 特征字段
	Title       []*FeatureRule            `ruleType:"array"`
	Header      []*FeatureRule            `ruleType:"array"`
	Cookie      []*FeatureRule            `ruleType:"array"`
	MetaTag     map[string][]*FeatureRule `ruleType:"map"`
	HeaderField map[string][]*FeatureRule `ruleType:"map"`
	CookieField map[string][]*FeatureRule `ruleType:"map"`
	Body        []*FeatureRule            `ruleType:"array"`
	// 缓存反射值和类型
	reflectType  reflect.Type
	reflectValue reflect.Value
}

type MatchedFeature struct {
	Feature  *Feature
	Versions []string
	Proofs   []string
}

func (rule *FeatureRule) getRegexp() *regexp.Regexp {
	if rule.regexp == nil {
		rule.regexp = regexp.MustCompile(rule.Regexp)
	}
	return rule.regexp
}

func (rule *FeatureRule) getVersionStock() *regexp.Regexp {
	if rule.versionStock == nil {
		rule.versionStock = regexp.MustCompile(rule.VersionStock)
	}
	return rule.versionStock
}

// 匹配内容，返回是否匹配成功和版本号
func (rule *FeatureRule) MatchContent(content string) (bool, []string, []string) {
	if len(content) == 0 {
		return false, nil, nil
	}
	var matched bool
	var versions []string
	var proofs []string

	if len(rule.Md5) > 0 {

	} else {
		var matchRe = rule.getRegexp()
		if matchRe == nil {
			return false, nil, nil
		}
		matched = matchRe.MatchString(content)

		if matched {
			proofs = append(proofs, fmt.Sprintf(":rule: %s :match: %s", rule.Regexp, matchRe.FindString(content)))
		}

		versionStockRe := rule.getVersionStock()
		if matched && versionStockRe != nil && len(rule.Version) > 0 {
			stocks := versionStockRe.FindAllStringSubmatch(content, -1)
			for _, stock := range stocks {
				version := rule.Version
				for i, split := range stock[1:] {
					version = strings.Replace(version, fmt.Sprintf("\\%d", i+1), split, -1)
				}
				versions = append(versions, strings.TrimSpace(version))
			}
		}
	}

	return matched, versions, proofs
}

func uniq(arr []string) []string {
	valueMap := make(map[string]bool)
	for _, v := range arr {
		valueMap[v] = true
	}
	var newArr []string
	for value, _ := range valueMap {
		newArr = append(newArr, value)
	}
	return newArr
}

func (feature *Feature) Type() reflect.Type {
	if feature.reflectType == nil {
		feature.reflectType = reflect.TypeOf(*feature)
	}
	return feature.reflectType
}

func (feature *Feature) Value() reflect.Value {
	if !feature.reflectValue.IsValid() {
		feature.reflectValue = reflect.ValueOf(*feature)
	}
	return feature.reflectValue
}

func (feature *Feature) MatchResponseData(response *ResponseData) *MatchedFeature {
	var matched = false
	var versions []string
	var proofs []string

	responseValue := reflect.ValueOf(*response)
	featureType := feature.Type()
	featureValue := feature.Value()

	for i := 0; i < featureType.NumField(); i++ {
		featureFieldValue := featureValue.Field(i)
		if featureValue.Field(i).IsZero() {
			continue
		}
		featureFieldType := featureType.Field(i)
		ruleType := featureFieldType.Tag.Get("ruleType")
		if ruleType != "array" && ruleType != "map" {
			continue
		}
		responseFieldValue := responseValue.FieldByName(featureFieldType.Name)
		if responseFieldValue.IsZero() {
			continue
		}

		if ruleType == "array" {
			for _, rule := range featureFieldValue.Interface().([]*FeatureRule) {
				currentMatched, currentVersions, currentProofs := rule.MatchContent(responseFieldValue.Interface().(string))
				matched = matched || currentMatched
				if len(currentVersions) > 0 {
					versions = append(versions, currentVersions...)
				}
				if len(currentProofs) > 0 {
					proofs = append(proofs, currentProofs...)
				}
			}
		} else if ruleType == "map" {
			for key, rules := range featureFieldValue.Interface().(map[string][]*FeatureRule) {
				for _, rule := range rules {
					currentMatched, currentVersions, currentProofs := rule.MatchContent(responseFieldValue.Interface().(map[string]string)[key])
					matched = matched || currentMatched
					if len(currentVersions) > 0 {
						versions = append(versions, currentVersions...)
					}
					if len(currentProofs) > 0 {
						proofs = append(proofs, currentProofs...)
					}
				}
			}
		}
	}

	if matched {
		return &MatchedFeature{
			Feature:  feature,
			Versions: uniq(versions),
			Proofs:   uniq(proofs),
		}
	} else {
		return nil
	}
}

func init() {
	// 初始化正则匹配表达式
	for _, feature := range Features {
		featureType := feature.Type()
		featureValue := feature.Value()

		for i := 0; i < featureType.NumField(); i++ {
			featureFieldValue := featureValue.Field(i)
			if featureValue.Field(i).IsZero() {
				continue
			}
			featureFieldType := featureType.Field(i)
			ruleType := featureFieldType.Tag.Get("ruleType")
			if ruleType != "array" && ruleType != "map" {
				continue
			}

			if ruleType == "array" {
				for _, rule := range featureFieldValue.Interface().([]*FeatureRule) {
					rule.regexp = regexp.MustCompile(rule.Regexp)
				}
			} else if ruleType == "map" {
				for _, rules := range featureFieldValue.Interface().(map[string][]*FeatureRule) {
					for _, rule := range rules {
						rule.regexp = regexp.MustCompile(rule.Regexp)
					}
				}
			}
		}
	}
	err := recover()
	if err != nil {
		log.Panic(err)
	}
}
