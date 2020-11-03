const wappalyzer = require("./fingers/wappalyzer.json");
const tide = require("./fingers/tide.json");
const fofa = require("./fingers/fofa.json");
const dayu = require("./fingers/dayu.json");
const whatweb = require("./fingers/whatweb.json");
const gwhatweb = require("./fingers/gwhatweb.json");
const custom = require("./fingers/custom.json")
const fs = require("fs");


class Feature {
  constructor(){
    this.Name = "";
    this.From = "";
    this.Path = "";
    this.Types = [];
    this.Implies = [];
    this.ManufacturerName = "";
    this.ManufacturerUrl = "";
    this.Title = [];
    this.Header = [];
    this.Cookie = [];
    this.Body = [];
    this.MetaTag = {};
    this.HeaderField = {};
    this.CookieField = {};
  }
}

class FeatureRule {
  constructor(reg, md5) {
    this.Regexp = reg || ""; // 用于匹配特征的正则
    this.Md5 = md5 || ""; // 用于匹配的md5
    this.VersionStock = ""; // 用于匹配版本号的正则，默认为Regexp
    this.Version = ""; // 用于生成版本号的规则
  }
}

function parseGWhatweb() {
  return gwhatweb.map(({ url, re, name, md5 }) => {
    const feature = new Feature();
    feature.Path = url;
    feature.From = "gwhatweb";
    feature.Name = name;
    feature.Types = ["CMS"];
    if (re) {
      feature.Body.push(new FeatureRule(re, ""));
    }
    if (md5) {
      feature.Body.push(new FeatureRule("", md5));
    }

    return feature;
  });
}

function parseDayu() {
  return dayu.map(item => {
    const {
      program_name: name,
      url,
      manufacturerName,
      manufacturerUrl,
      recognition_content: md5
    } = item;
    const feature = new Feature();
    feature.Types = ["CMS"];
    feature.From = "dayu";
    feature.Name = name;
    feature.Path = url;
    feature.ManufacturerName = manufacturerName;
    feature.ManufacturerUrl = manufacturerUrl;
    feature.Body.push(new FeatureRule("", md5));
    return feature;
  });
}

function parseTide() {
  return tide.map(item => {
    const { cms_name: name, path: url, options, match_pattern } = item;
    const feature = new Feature();
    feature.From = "tide";
    feature.Types = ["CMS"];
    feature.Name = name;
    feature.Path = url;

    if (options === "md5") {
      feature.Body.push(new FeatureRule("", match_pattern));
    } else if (options === "keyword") {
      feature.Body.push(new FeatureRule(match_pattern.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), ""));
    }

    return feature;
  });
}

function parseFofa() {
  return fofa.map(({ name, keys }) => {
    const feature = new Feature();
    feature.From = "fofa";
    feature.Name = name;
    feature.Path = "/";

    const keyItems = keys.split(/\|\|/g);

    keyItems.forEach(key => {
      if (/(^\(.*\)$)/.test(key)) {
        key = key.replace(/(^\(|\)$)/g, "");
      }
      key = key.trim();
      if (/\s&&\s/.test(key)) {
        console.log(`无法处理: ${key}`);
      } else {
        let equal = key.indexOf("=");
        if (equal <= 0) {
          return console.log(`错误的格式: ${key} :${keys}: ${name}`);
        }
        let pre = key.slice(0, equal).trim();
        let post = key
          .slice(equal + 1)
          .trim()
          .replace(/(^"|"$)/g, "")
          .replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        if (pre === "title") {
          feature.Title.push(new FeatureRule(post));
        } else if (pre === "header") {
          feature.Header.push(new FeatureRule(post));
        } else if (pre === "cookie") {
          feature.Cookie.push(new FeatureRule(post));
        } else if (pre === "body") {
          feature.Body.push(new FeatureRule(post));
        } else if (pre === "server") {
          feature.HeaderField["server"] = [new FeatureRule(post)];
        } else {
          console.log(`错误的格式: ${key}: ${keys}: ${name}`);
        }
      }
    });

    if (!keys) {
      console.log(`无指纹信息: ${name}: ${keys}`);
    }

    return feature;
  });
}

function parseWappalayzer() {
  const features = [];
  const { apps, categories } = wappalyzer;

  function parseRule(rulestr) {
    rulestr = rulestr.replace(/\\;confidence:\d{0,3}/g, "");
    const versionExec = /\\;version:(.+)$/g.exec(rulestr);
    let version;
    if (versionExec != null) {
      version = versionExec[1];
      rulestr = rulestr.replace(/\\;version:.+$/g, "");
    }
    const rule = new FeatureRule(rulestr, "");
    if (version) {
      rule.VersionStock = rulestr;
      rule.Version = version;
    }
    return rule;
  }

  for (let name in apps) {
    const {
      cats = [],
      headers,
      html = [],
      script = [],
      cookies,
      implies = [],
      meta
    } = apps[name];
    const feature = new Feature();
    feature.Name = name;
    feature.From = "wappalayzer";
    feature.Types = cats.map(cat => categories[cat].name);
    feature.Implies = []
      .concat(implies)
      .map(i => i.replace(/\\;confidence:.*$/, ""));

    for (let field in cookies) {
      feature.CookieField[field] = [parseRule(cookies[field])];
    }
    for (let field in headers) {
      feature.HeaderField[field] = [parseRule(headers[field])];
    }
    for (let item of [].concat(html)) {
      feature.Body.push(parseRule(item));
    }
    for (let field in meta) {
      feature.MetaTag[field] = [parseRule(meta[field])];
    }
    for (let item of [].concat(script)) {
      feature.Body.push(parseRule(item.replace(/^\^/, "")));
    }
    features.push(feature);
  }

  return features;
}

function parseWhatweb() {
  const features = [];
  for (let name in whatweb) {
    for (let m of whatweb[name].matches) {
      const feature = new Feature();
      feature.Name = name;
      feature.From = "whatweb";
      feature.Path = m.url || "/";
      if ([m.regexp, m.text, m.md5].filter(Boolean).length > 1) {
        console.log(">>>>> more 1", m.regexp, m.md5, m.text);
      }

      const parseRegexp = str => {
        if (/^\(\?-mix:(.*)\)$/.test(str)) {
          return /^\(\?-mix:(.*)\)$/.exec(str)[1];
        } else if (/\(\?m-ix:(.*)\)$/.test(str)) {
          return /\(\?m-ix:(.*)\)$/.exec(str)[1];
        } else if (/\(\?i-mx:(.*)\)$/.test(str)) {
          return /\(\?i-mx:(.*)\)$/.exec(str)[1];
        }else if(/\(\?mi-x:(.*)\)$/.test(str)){
          return /\(\?mi-x:(.*)\)$/.exec(str)[1];
        } else {
          return str;
        }
      };
      if (m.regexp) {
        m.regexp = parseRegexp(m.regexp);
      }
      if (m.version) {
        m.version = parseRegexp(m.version);
      }
      if(m.text){
        m.text = m.text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
      }

      const rule = new FeatureRule(m.regexp || m.text || m.version, m.md5);

      if (m.version) {
        rule.VersionStock = m.version;
        rule.Version = "\\1";
      }

      if (!m.search || m.search === "body") {
        feature.Body.push(rule);
      } else if (m.search === "headers") {
        feature.Header.push(rule);
      } else if (/headers\[(.+)\]/.test(m.search)) {
        const key = /headers\[(.+)\]/.exec(m.search)[1];
        feature.HeaderField[key] = [rule];
      }
      features.push(feature);
    }
  }

  return features;
}

function parseWebeye() {
  const origin = fs.readFileSync("./fingers/webeye.txt").toString();
  const lines = origin.split("\n").filter(i => /^[^;\[\s]/.test(i));
  const features = [];
  for (line of lines) {
    const re = /^(\w+?):(.+?)\|(.+?)\|(.+?)\|(.+)$/g;
    if (line.match(re)) {
      const splits = re.exec(line);
      // splits[4]=splits[4].replace(/\\/g, "\\$&")
      const feature = new Feature();
      feature.Name = splits[2];
      feature.From = "webeye";
      feature.Types = [splits[1]];
      if (splits[3] === "url") {
        feature.Path = splits[4];
      } else {
        feature.Path = "/";
      }
      if (splits[3] === "headers") {
        feature.HeaderField[splits[4]] = [new FeatureRule(splits[5])];
      } else {
        feature.Body.push(new FeatureRule(splits[5]));
      }
      features.push(feature);
    } else {
      console.log(`解析webeye错误: ${line}`);
    }
  }
  return features;
}

function uniqArray(arr) {
  return Array.from(new Set(arr.reduce((t, c) => t.concat(c), []))).filter(
    Boolean
  );
}

function uniqRules(rules, clear = true) {
  const ruleMap = {};
  uniqArray(rules).forEach(rule => {
    const key = [rule.Regexp, rule.Md5, rule.VersionStock, rule.Version]
      .map(i => {
        return (i || "").toString().toLowerCase();
      })
      .join("");
    if (!key.trim() && clear) {
      return;
    }
    ruleMap[key] = rule;
  });
  return Object.values(ruleMap);
}

function uniq(features) {
  const featureMap = {};
  for (let feature of features) {
    feature.Name = feature.Name.trim()
    if (!feature.Name){
      continue
    }
    if(!/^\//.test(feature.Path)){
      feature.Path = "/"+feature.Path
    }
    const key = `${feature.Name.replace(
      /[^0-9a-zA-Z\u4e00-\u9fa5]/g,
      ""
    ).toLowerCase()}_${feature.Path}`;
    if (!featureMap[key]) {
      featureMap[key] = [];
    }
    featureMap[key].push(feature);
  }
  for (let key in featureMap) {
    const gfs = featureMap[key];
    const feature = new Feature();
    feature.Name = gfs[0].Name;
    feature.Path = gfs[0].Path;

    feature.From = uniqArray(gfs.map(i => i.From));
    feature.Types = uniqArray(gfs.map(i => i.Types));
    feature.Implies = uniqArray(gfs.map(i => i.Implies));
    feature.ManufacturerName = uniqArray(gfs.map(i => i.ManufacturerName));
    feature.ManufacturerUrl = uniqArray(gfs.map(i => i.ManufacturerUrl));
    feature.Title = uniqRules(gfs.map(i => i.Title));
    feature.Header = uniqRules(gfs.map(i => i.Header));
    feature.Cookie = uniqRules(gfs.map(i => i.Cookie));
    feature.Body = uniqRules(gfs.map(i => i.Body));

    let tmpKeys;
    tmpKeys = uniqArray(gfs.map(f => Object.keys(f.MetaTag)));
    for (let rkey of tmpKeys) {
      feature.MetaTag[rkey] = uniqRules(
        gfs.map(i => i.MetaTag[rkey] || []),
        false
      );
    }
    tmpKeys = uniqArray(gfs.map(f => Object.keys(f.HeaderField)), false);
    for (let rkey of tmpKeys) {
      feature.HeaderField[rkey] = uniqRules(
        gfs.map(i => i.HeaderField[rkey] || [])
      );
    }
    tmpKeys = uniqArray(gfs.map(f => Object.keys(f.CookieField)));
    for (let rkey of tmpKeys) {
      feature.CookieField[rkey] = uniqRules(
        gfs.map(i => i.CookieField[rkey] || []),
        false
      );
    }
    featureMap[key] = feature;
  }

  return Object.values(featureMap).sort(function (a, b) {
    if( a.Name.toLowerCase() > b.Name.toLowerCase()){
      return 1
    }else if(a.Name.toLowerCase()<b.Name.toLowerCase()){
      return -1
    }else {
      if( a.Path.toLowerCase() > b.Path.toLowerCase()){
        return 1
      }else if(a.Path.toLowerCase()<b.Path.toLowerCase()){
        return -1
      }else {
        console.log("error path", a.Name, a.Path)
        return  0
      }
    }
  });
}

function transformRule(rule) {
  const lines = Object.keys(rule)
    .filter(key => key === "Regexp" || rule[key])
    .map(key => `${key}: ${JSON.stringify(rule[key].toString())}`);
  return `&FeatureRule{\n${lines.join(",\n")},\n}`;
}

function transformRuleArray(rules) {
  const lines = rules.map(transformRule);
  const tail = lines.length > 0 ? "," : "";
  return `[]*FeatureRule{\n${lines.join(",\n")}${tail}\n}`;
}

function transformRuleMap(ruleMap) {
  return `map[string][]*FeatureRule{\n${Object.keys(ruleMap)
    .map(key => `"${key}": ${transformRuleArray(ruleMap[key])}`)
    .join(",\n")},\n}`;
}

function transformFeature(feature, id) {
  try {
    const lines = [`ID:${id}`];
    for (let key in feature) {
      if (["Name", "Path"].includes(key)) {
        lines.push(`${key}: "${feature[key]}"`);
      } else if (
        [
          "From",
          "Types",
          "Implies",
          "ManufacturerName",
          "ManufacturerUrl"
        ].includes(key)
      ) {
        if (feature[key].length > 0) {
          lines.push(
            `${key}: []string{\n${feature[key]
              .map(i => JSON.stringify(i))
              .join(",\n")},\n}`
          );
        }
      } else if (
        ["Title", "Header", "Cookie", "Body", "ManufacturerUrl"].includes(key)
      ) {
        if (feature[key].length > 0) {
          lines.push(`${key}: ${transformRuleArray(feature[key])}`);
        }
      } else if (["MetaTag", "HeaderField", "CookieField"].includes(key)) {
        if (Object.keys(feature[key]).length > 0) {
          lines.push(`${key}: ${transformRuleMap(feature[key])}`);
        }
      }
    }
    return `&Feature{\n${lines.join(",\n")},\n}`;
  } catch (err) {
    console.log(feature);
    console.log(err);
    process.exit();
  }
}

function transformFeatureArray(features) {
  return `[]*Feature{\n${features.map(transformFeature).join(",\n")},\n}`;
}

function checkMostLike(features){
  const namePathMap = {}
  for(let feature of features){
    const trimed = `${feature.Name
      .replace(/[^0-9a-zA-Z\u4e00-\u9fa5]/g, "")
      .toLowerCase()}_${feature.Path}`;
    if(!namePathMap[trimed]){
      namePathMap[trimed] = []
    }
    namePathMap[trimed].push(feature.Name)
  } 
  let count= 1
  for(let key in namePathMap){
    if(namePathMap[key].length > 1){
      console.log(`${count}: ${key}: ${JSON.stringify(namePathMap[key])}`)
      count++;
    }
  }
}

function main() {
  let features = [];
  features.push(
    ...parseGWhatweb(),
    ...parseDayu(),
    ...parseTide(),
    ...parseFofa(),
    ...parseWappalayzer(),
    ...parseWhatweb(),
    ...parseWebeye(),
    ...custom
  );
  features = uniq(features);
  console.log(`features count: ${features.length}`);
  checkMostLike(features)
  fs.writeFileSync(
    "../appscan/source.go",
    `package appscan\n\nvar Features = ${transformFeatureArray(features)}`
  );
}

main();
