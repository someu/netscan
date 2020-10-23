const wappalyzer = require("./fingers/wappalyzer.json");
const tide = require("./fingers/tide.json");
const fofa = require("./fingers/fofa.json");
const dayu = require("./fingers/dayu.json");
const whatweb = require("./fingers/whatweb.json");
const gwhatweb = require("./fingers/gwhatweb.json");
const fs = require("fs");
const _ = require("loadsh");

// 特征
class FeatureGroup {
  Path = "";
  Features = [];
}

class Feature {
  Name = "";
  From = "";
  Path = "";
  Types = [];
  Implies = [];
  ManufacturerName = "";
  ManufacturerUrl = "";
  Title = [];
  Header = [];
  Cookie = [];
  Body = [];
  MetaTag = {};
  HeaderField = {};
  CookieField = {};
}

class FeatureRule {
  Regexp = "";
  Md5 = "";
  Version = "";
  constructor(reg, md5) {
    this.Regexp = reg;
    this.Md5 = md5;
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
  return dayu.map((item) => {
    const {
      program_name: name,
      url,
      manufacturerName,
      manufacturerUrl,
      recognition_content: md5,
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
  return tide.map((item) => {
    const { cms_name: name, path: url, options, match_pattern } = item;
    const feature = new Feature();
    feature.From = "tide";
    feature.Types = ["CMS"];
    feature.Name = name;
    feature.Path = url;

    if (options === "md5") {
      feature.Body.push(new FeatureRule("", match_pattern));
    } else if (options === "keyword") {
      feature.Body.push(new FeatureRule(match_pattern, ""));
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

    keyItems.forEach((key) => {
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
          feature.Title.push(new FeatureRule(post, ""));
        } else if (pre === "header") {
          feature.Header.push(new FeatureRule(post, ""));
        } else if (pre === "cookie") {
          feature.Cookie.push(new FeatureRule(post, ""));
        } else if (pre === "body") {
          feature.Body.push(new Feature(post, ""));
        } else if (pre === "server") {
          feature.HeaderField["server"].push(new FeatureRule(post, ""));
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
    rulestr = rulestr.replace(/\\;confidence:\d+$/, "");
    const versionExec = /\\;version:(.+)$/g.exec(rulestr);
    let version
    if (versionExec != null) {
      version = versionExec[0];
      rulestr = rulestr.replace(/\\;version:.+$/g, "");
    }
    const rule = new FeatureRule(Re, "");
    if (version) {
      rule.Version = version
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
      meta,
    } = apps[name];
    const feature = new Feature();
    feature.Name = name;
    feature.From = "wappalayzer";
    feature.Types = cats.map((cat) => categories[cat].name);
    feature.Implies = implies;

    for (let field in cookies) {
      feature.Cookie[field] = parseRule(cookies[field]);
    }
    for (let field in headers) {
      feature.Header[field] = parseRule(headers[field]);
    }
    for (let item of [].concat(html)) {
      feature.Body.push(parseRule(item))
    }
    for (let field in meta) {
      feature.MetaTag[field] = parseRule(meta[field]);
    }
    for (let item of [].concat(script)) {
      feature.Body.push(parseRule(item.replace(/^\^/, "")));
    }
    features.push(feature)
  }

  return features;
}
