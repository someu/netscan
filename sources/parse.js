const features = require("./features.json");
const fs = require("fs");

function keyToUpperCase(target) {
  if (target instanceof Object) {
    if (Array.isArray(target)) {
      return target.map(keyToUpperCase);
    } else {
      const newTarget = {};
      for (let key in target) {
        newTarget[key[0].toUpperCase() + key.slice(1)] = keyToUpperCase(
          target[key]
        );
      }
      return newTarget;
    }
  } else {
    return target;
  }
}

function toString(target) {
  if (target instanceof Object) {
    if (Array.isArray(target)) {
      return `{\n${target.map(toString).join(",\n")},\n}`;
    } else {
      return `{\n${Object.keys(target)
        .map((key) => {
          return `${key}:${toString(target[key])},`;
        })
        .join("\n")}\n}`;
    }
  } else {
    return `${JSON.stringify(target)}`;
  }
}

const formatedFeatures = keyToUpperCase(features);

fs.writeFileSync(
  "./source.go",
  `package appscan

var Features = []*Feature${toString(formatedFeatures, "")}`
);
