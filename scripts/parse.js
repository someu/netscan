// 网站title
// 网站header
// 网站cookie
// 主页body 正则
// 特定页面 正则
// 特定页面 md5
// 暗示
// 版本提取
// class Rule {
//   name = String;
//   rules = [
//     {
//       url: "/index.html",
//       md5: "asdads",
//       regexp: "asdad",
//     },
//   ];
// }

// class Rule {
//   type = "regexp|md5";
//   value = "12";
// }
// rules = {
//   content: {
//     title: { type: "re", value: "as" },
//     body: { type: "re", value: "asas" },
//   },
//   header: {
//     xportby: { type: "re", value: "as" },
//   },
//   cookie: {
//     xportby: { type: "re", value: "as" },
//   },
//   pages: {
//     "/asas": { type: "re", value: "as" },
//     "/asa2s": { type: "md5", value: "as" },
//   },
// };

// class Target {
//   titles = [];
//   pages = {
//     "/": {
//       title: "",
//       header: {},
//       cookie: {},
//       md5: "",
//       body: "",
//     },
//   };
// }

// class App {
//   name = "DEDE";

//   pageRules = {
//     "/": {
//       title: { type: "re", value: "as" },
//       header: {},
//       cookie: {},
//       md5: "",
//       body: "",
//     },
//   };
//   imples = [];
// }
const wappalyzer = require("./fingers/wappalyzer.json");
const tide = require("./fingers/tide.json");
const fofa = require("./fingers/fofa.json");
const dayu = require("./fingers/dayu.json");
const whatweb = require("./fingers/whatweb.json");
const gwhatweb = require("./fingers/gwhatweb.json");
const fs = require("fs");
const _ = require("loadsh");

class Rule {
  constructor(type, value, version) {
    if (type === Re) {
      this.regexp = value;
    } else if (type === Md5) {
      this.md5 = value;
    }
    // if (version) {
    //   // [if,else],
    //   this.version = {
    //     type: "value",
    //     value: ["qw", "\1"],
    //   };
    //   this.version = {
    //     type: "ifthen",
    //     value: [1, "\1", "\2"],
    //   };
    // }
  }
}

function clearObj(obj) {
  if (!obj) {
    return null;
  }
  if (typeof obj === "object") {
    if (Array.isArray(obj)) {
      const r = obj.map(clearObj).filter((i) => !_.isEmpty(i));
      if (r.length) {
        return r;
      } else {
        return null;
      }
    } else {
      let no = {};
      for (let k in obj) {
        const i = clearObj(obj[k]);
        if (!_.isEmpty(i)) {
          no[k] = i;
        }
      }
      return _.isEmpty(no) ? null : no;
    }
  } else {
    return obj;
  }
}

function combineItem(o1, o2) {
  if (Array.isArray(o1) && Array.isArray(o2)) {
    return _.uniqBy(o1.concat(o2), _.isEqual);
  } else if (typeof o1 === "object" && typeof o2 === "object") {
    let newO = o1;
    for (let key in o2) {
      if (!newO[key]) {
        newO[key] = o2[key];
      } else {
        newO[key] = combineItem(newO[key], o2[key]);
      }
    }

    return newO;
  } else if (o1 === o2) {
    return o1;
  } else {
    console.log(`combine error: ${o1} : ${o2}`);
  }
}

function combine(apps) {
  let appObj = {};
  for (let app of apps) {
    if (!appObj[app.name]) {
      appObj[app.name] = app;
    } else {
      appObj[app.name] = combineItem(appObj[app.name], app);
    }
  }

  return Object.values(appObj);
}

class App {
  constructor(name, from) {
    this.name = name;
    this.from = from;
    this.pageRules = {};
  }
  createPageRule(page) {
    if (!this.pageRules[page]) {
      this.pageRules[page] = {
        title: [],
        header: [],
        cookie: [],
        metaTag: {},
        headerField: {},
        cookieField: {},
        md5: [],
        body: [],
      };
    }
  }
  valueOf() {
    let pageRules = {};
    for (let key in this.pageRules) {
      let pageRule = this.pageRules[key];
      pageRules[key] = _.pickBy(pageRule, (i) => !_.isEmpty(i));
    }
    const value = {
      name: this.name,
      from: [this.from],
      rules: pageRules,
    };
    if (this.implies) {
      value.implies = this.implies;
    }
    return value;
  }
}

const Re = "regexp";
const Md5 = "md5";

function checkMd5(md5) {
  if (!/[a-zA-Z0-9]{32}/.test(md5)) {
    console.log(`错误的md5: ${md5}`);
  }
}

function main() {
  const apps = [];

  (function parseGWhatweb() {
    console.log("gwhatweb>>>>");
    gwhatweb.forEach(({ url, re, name, md5 }) => {
      const app = new App(name, "gwhatweb");
      app.createPageRule(url);
      if (re) {
        app.pageRules[url].body.push(new Rule(Re, re));
      }
      if (md5) {
        checkMd5(md5);
        app.pageRules[url].body.push(new Rule(Md5, md5));
      }
      if (!re && !md5) {
        console.log(`无指纹信息: ${name}: ${url}`);
      }

      apps.push(app.valueOf());
    });
  })();

  (function parseDayu() {
    console.log("dayu>>>>");
    dayu.forEach(
      ({ program_name: name, url, regexp, recognition_content: md5 }) => {
        const app = new App(name, "dayu");
        app.createPageRule(url);
        if (md5.length != 32) {
          //   console.log(`md5错误: ${md5}: ${md5.length}`);
          app.pageRules[url].body.push(new Rule(Re, md5));
        } else {
          checkMd5(md5);
          app.pageRules[url].body.push(new Rule(Md5, md5));
        }

        if (!md5) {
          console.log(`无指纹信息: ${name}: ${url}`);
        }

        apps.push(app.valueOf());
      }
    );
  })();

  (function parsetide() {
    console.log("tide>>>>");
    tide.forEach(({ cms_name: name, path: url, options, match_pattern }) => {
      const app = new App(name, "tide");
      app.createPageRule(url);
      if (options === "keyword") {
        //   console.log(`md5错误: ${md5}: ${md5.length}`);
        app.pageRules[url].body.push(new Rule(Re, match_pattern));
      } else if (options === "md5") {
        checkMd5(match_pattern);
        app.pageRules[url].body.push(new Rule(Md5, match_pattern));
      }

      if (!match_pattern) {
        console.log(`无指纹信息: ${name}: ${url}`);
      }

      apps.push(app.valueOf());
    });
  })();

  (function parsefofa() {
    console.log("fofa>>>>");
    fofa.forEach(({ name, keys }) => {
      const app = new App(name, "fofa");
      app.createPageRule("/");

      const keyItems = keys.split(/\|\|/g);
      let wrong = false;

      keyItems.forEach((key) => {
        if (/(^\(.*\)$)/.test(key)) {
          key = key.replace(/(^\(|\)$)/g, "");
        }
        key = key.trim();
        if (/\s&&\s/.test(key)) {
          wrong = true;
          console.log(`无法处理: ${key}`);
        } else {
          let equal = key.indexOf("=");
          if (equal <= 0) {
            wrong = true;
            return console.log(`错误的格式: ${key} :${keys}: ${name}`);
          }
          let pre = key.slice(0, equal).trim();
          let post = key
            .slice(equal + 1)
            .trim()
            .replace(/(^"|"$)/g, "")
            .replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
          if (["title", "header", "cookie", "body"].includes(pre)) {
            app.pageRules["/"][pre].push(new Rule(Re, post));
          } else if (pre === "server") {
            if (!app.pageRules["/"].headerField["server"]) {
              app.pageRules["/"].headerField["server"] = [];
            }
            app.pageRules["/"].headerField["server"].push(new Rule(Re, post));
          } else {
            wrong = true;
            console.log(`错误的格式: ${key}: ${keys}: ${name}`);
          }
        }
      });

      if (!keys) {
        wrong = true;
        console.log(`无指纹信息: ${name}: ${keys}`);
      }

      !wrong && apps.push(app.valueOf());
    });
  })();

  (function parsewapplayzer() {
    console.log("wapplayzer>>>>");
    for (let key in wappalyzer.apps) {
      const app = new App(key, "wappalyzer");
      app.createPageRule("/");
      const {
        cookies,
        implies,
        html = [],
        headers,
        script = [],
        meta,
      } = wappalyzer.apps[key];

      function parseRule(rulestr) {
        rulestr = rulestr.replace(/\\;confidence:\d+$/, "");
        const versionStr = _.get(/\\;version:(.+)$/g.exec(rulestr), 1);
        rulestr = rulestr.replace(/\\;version:.+$/g, "");
        const rule = new Rule(Re, rulestr);
        let version;
        if (versionStr) {
          console.log(">>>>", versionStr);
          if (versionStr.includes("?") && versionStr.includes(":")) {
            const execs = /(.*)\?([^:]*):(.*)/g.exec(versionStr);
            rule.version = {
              if: execs[1],
              then: execs[2],
              else: execs[3],
            };
          } else {
            // const execs = /(.*)\?([^:]*):(.*)/g.exec(versionStr);
            rule.version = {
              value: versionStr,
            };
          }
          // console.log(">>>>>>>>>> ", versionStr);
          // if (/\\;version:(.+)\\1$/.test(versionStr)) {
          //   const execs = /\\;version:(.+)\\1/g.exec(versionStr);
          //   if (execs) {
          //     version = [execs[1], "", "", versionStr];
          //   }
          // } else if (/\\;version:\\1\??(.*):?(.*)/.test(versionStr)) {
          //   const execs = /\\;version:\\1\??([^:]*):?(.*)/g.exec(versionStr);
          //   if (execs) {
          //     version = ["", execs[1], execs[2], versionStr];
          //   }
          // }
        }

        if (version) {
          rule.version = version;
        }
        return rule;
      }

      for (let field in cookies) {
        app.pageRules["/"].cookie[field] = parseRule(cookies[field]);
      }
      for (let field in headers) {
        app.pageRules["/"].header[field] = parseRule(headers[field]);
      }
      for (let item of [].concat(html)) {
        if (!app.pageRules["/"].body) {
          app.pageRules["/"].body = [];
        }
        app.pageRules["/"].body.push(parseRule(item));
      }
      for (let field in meta) {
        app.pageRules["/"].metaTag[field] = parseRule(meta[field]);
      }
      for (let item of [].concat(script)) {
        if (!app.pageRules["/"].body) {
          app.pageRules["/"].body = [];
        }
        app.pageRules["/"].body.push(parseRule(item.replace(/^\^/, "")));
      }
      if (implies && implies.length) {
        app.implies = implies;
      }
      apps.push(app.valueOf());
    }
  })();

  (function parsewhatweb() {
    console.log("whatweb>>>>");
    for (let name in whatweb) {
      const app = new App(name, "whatweb");
      const matches = whatweb[name].matches;
      for (let m of matches) {
        if (!m.url) {
          m.url = "/";
        }
        if (!app.pageRules[m.url]) {
          app.createPageRule(m.url);
        }
        const pageRule = app.pageRules[m.url];
        const parseRegexp = (str) => {
          if (/^\(\?-mix:(.*)\)$/.test(str)) {
            return _.get(/^\(\?-mix:(.*)\)$/.exec(str), 1);
          } else if (/\(\?m-ix:(.*)\)$/.test(str)) {
            return _.get(/\(\?m-ix:(.*)\)$/.exec(str), 1);
          } else if (/\(\?i-mx:(.*)\)$/.test(str)) {
            return _.get(/\(\?i-mx:(.*)\)$/.exec(str), 1);
          } else {
            return str;
          }
        };
        if ([m.regexp, m.text, m.md5].filter(Boolean).length > 1) {
          console.log(">>>>> more 1", m.regexp, m.md5, m.text);
        }
        if (m.regexp) {
          m.regexp = parseRegexp(m.regexp);
        }
        let rule = new Rule();
        let vr = parseRegexp(m.version);
        if (m.regexp || m.text || vr) {
          rule.regexp = m.regexp || m.text || vr;
        }
        if (m.md5) {
          rule.md5 = m.md5;
        }
        if (vr) {
          rule.version = { match: vr };
        }
        if (!m.search || m.search === "body") {
          pageRule.body.push(rule);
        } else if (m.search === "headers") {
          pageRule.header.push(rule);
        } else if (/headers\[(.+)\]/.test(m.search)) {
          const headerKey = _.get(/headers\[(.+)\]/.exec(m.search), 1);
          if (!pageRule.headerField[headerKey]) {
            pageRule.headerField[headerKey] = [];
          }
          pageRule.headerField[headerKey].push(rule);
        }
      }

      apps.push(app.valueOf());
    }
  })();

  function transform(apps) {
    let t = {};
    for (let app of apps) {
      for (let path in app.rules) {
        if (!/^\//.test(path)) {
          path = "/" + path;
        }
        if (!t[path]) {
          t[path] = [];
        }
        t[path].push({
          name: app.name,
          ...app.rules[path],
        });
      }
    }
    let tt = [];
    for (let p in t) {
      tt.push({
        path: p,
        rules: t[p],
      });
    }
    return tt.sort((a, b) => -a.rules.length + b.rules.length);
  }

  const ta = transform(
    combine(clearObj(apps).filter((i) => !_.isEmpty(i.rules)))
  );

  ta.forEach((t) => {
    t.rules.forEach((rule) => {
      [rule.headerField, rule.metaTag, rule.cookieField].forEach((g) => {
        if (g) {
          for (let k in g) {
            v = g[k];
            delete g[k];
            g[k.toLowerCase()] = v;
          }
        }
      });
    });
  });

  fs.writeFileSync("./total.json", JSON.stringify(ta, null, 2));
  fs.writeFileSync(
    "./features.json",
    JSON.stringify(
      ta.map((i) => {
        delete i.from;
        return i;
      }),
      null,
      2
    )
  );

  function keyToUpperCase(target) {
    if (target instanceof Object) {
      if (Array.isArray(target)) {
        return target.map(keyToUpperCase);
      } else {
        const newTarget = {};
        for (let key in target) {
          if (target[key] instanceof Rule) {
            newTarget[key] = keyToUpperCase(target[key]);
          } else {
            newTarget[key[0].toUpperCase() + key.slice(1)] = keyToUpperCase(
              target[key]
            );
          }
        }
        return newTarget;
      }
    } else {
      return target;
    }
  }

  function toString(target, parentKey, isFeatureRuleItem) {
    let res;
    let isFeatureRuleItems
    if (target instanceof Object) {
      if (Array.isArray(target)) {
        res = `{\n${target.map(i=>(toString(i))).join(",\n")},\n}`;
      } else {
        res = `{\n${Object.keys(target)
          .map((key) => {
            let rkey = key;
            if (["MetaTag", "HeaderField", "CookieField"].includes(parentKey)) {
              rkey = `"${key}"`;
              isFeatureRuleItems = true
            }
            return `${rkey}:${toString(target[key], key, isFeatureRuleItems)},`;
          })
          .join("\n")}\n}`;
      }
    } else {
      res = `${JSON.stringify(target)}`;
    }

    const ArrKeys = [
      "Title",
      "Header",
      "Cookie",
       
      "Body",
    ];
    const MapKeys = [
      "MetaTag",
      "HeaderField",
      "CookieField",
    ]

    if (ArrKeys.includes(parentKey)) {
      return `[]*FeatureRuleItem${res}`
    }else if(MapKeys.includes(parentKey)){
      return `map[string][]*FeatureRuleItem${res}`
    }else if ("Rules" === parentKey){
      return `[]*FeatureRule${res}`
    }else if("Version" === parentKey){
      return `&FeatureVersion${res}`
    }else if (isFeatureRuleItems){
      return `[]*FeatureRuleItem\n${res},\n`
    }else if (isFeatureRuleItem){
      return  `&FeatureRuleItem${res}`
    }

    return res;
  }

  const features = ta.map((i) => {
    delete i.from;
    return i;
  });

  const formatedFeatures = keyToUpperCase(features);

  fs.writeFileSync(
    "../appscan/source.go",
    `package appscan
  
  var features = []*Feature${toString(formatedFeatures, "")}`
  );

  console.log(`app> ${ta.length}`);
}

main();
