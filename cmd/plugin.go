package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
	"sort"
)

var search string

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Print the plugins",
	Long:  `Print the plugins`,
	Run: func(cmd *cobra.Command, args []string) {
		ruleMap := map[string]int{}
		var searchRe *regexp.Regexp
		if len(search) > 0 {
			searchRe = regexp.MustCompile(search)
		}
		for _, feature := range globalScanner.FeatureCollection {
			for _, rule := range feature.Rules {
				if searchRe == nil || searchRe.MatchString(rule.Name) {
					ruleMap[rule.Name] += 1
				}
			}
		}

		type ruleType struct {
			name  string
			count int
		}

		var rules []ruleType

		for name, count := range ruleMap {
			rules = append(rules, ruleType{name: name, count: count})
		}
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].count >= rules[j].count
		})
		for index, rule := range rules {
			fmt.Printf("%d\t%d\t%s\n", index+1, rule.count, rule.name)
		}
	},
}

func init() {
	pluginCmd.Flags().StringVarP(&search, "search", "s", "", "filter rules")
	rootCmd.AddCommand(pluginCmd)
}
