process.env.NODE_ENV = "production";

const WebpackMerge = require("webpack-merge");
const webpackCommonConfig = require("./webpack.common");

module.exports = WebpackMerge.merge(webpackCommonConfig, {
  mode: "production",
  performance: {
    hints: "error",
    maxEntrypointSize: 1024000,
    maxAssetSize: 1024000
  }
});
