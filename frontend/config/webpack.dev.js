process.env.NODE_ENV = "development";

const path = require("path");
const WebpackMerge = require("webpack-merge");
const webpackCommonConfig = require("./webpack.common");

module.exports = WebpackMerge.merge(webpackCommonConfig, {
  mode: "development",
  devtool: "cheap-module-eval-source-map",
  devServer: {
    host: "0.0.0.0",
    contentBase: path.resolve(__dirname, "build"),
    hot: true,
    clientLogLevel: "none",
    historyApiFallback: true,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8899",
        changeOrigin: true,
        secure: false,
      },
    },
    headers: {
      "Access-Control-Allow-Origin": "*",
    },
  },
});
