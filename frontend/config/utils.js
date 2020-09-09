const path = require("path");

const basePaths = {
  appEntry: path.resolve(__dirname, "../src/index.js"),
  appHtml: path.resolve(__dirname, "../public/index.html"),
  output: path.resolve(__dirname, "../build"),
  public: path.resolve(__dirname, "../public"),
  nodeModules: path.resolve(__dirname, "../node_modules")
};

function getCssLoaders(frontLoaders = []) {
  return [
    "style-loader",
    {
      loader: "css-loader",
      options: {
        importLoaders: frontLoaders.length + 1
      }
    },
    {
      loader: "postcss-loader",
      options: {
        plugins: [require("autoprefixer")]
      }
    },
    ...frontLoaders
  ];
}

module.exports = {
  basePaths,
  getCssLoaders
};
