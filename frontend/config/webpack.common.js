const HtmlWebpackPlugin = require("html-webpack-plugin");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const { basePaths, getCssLoaders } = require("./utils");

module.exports = {
  entry: {
    app: basePaths.appEntry,
  },
  output: {
    publicPath: "/",
    path: basePaths.output,
    filename: "static/js/[name].[hash:8].js",
    chunkFilename: "static/js/[name].[chunkhash:8].chunk.js",
  },
  resolve: {
    extensions: [".js", ".jsx"],
    modules: [basePaths.nodeModules],
  },
  module: {
    rules: [
      {
        oneOf: [
          {
            test: /\.js(x)?$/,
            exclude: /node_modules/,
            use: [
              {
                loader: "babel-loader",
              },
            ],
          },
          {
            test: /\.css$/,
            use: getCssLoaders(),
          },
          {
            test: /\.less$/,
            use: getCssLoaders([
              {
                loader: "less-loader",
                options: { lessOptions: { javascriptEnabled: true } },
              },
            ]),
          },
          {
            test: /\.(bmp|png|svg|jpg|jpeg|gif)$/,
            use: {
              loader: "url-loader",
              options: {
                limit: 10000,
                name: "static/images/[name].[hash:8].[ext]",
              },
            },
          },
          {
            test: /\.(woff|woff2|eot|ttf|otf)$/,
            use: {
              loader: "file-loader",
              options: {
                name: "static/font/[name].[hash:8].[ext]",
              },
            },
          },
        ],
      },
    ],
  },
  plugins: [
    // new CopyWebpackPlugin({
    //   patterns: [
    //     {
    //       from: basePaths.public,
    //       to: basePaths.output,
    //     },
    //   ],
    // }),
    new HtmlWebpackPlugin({
      template: basePaths.appHtml,
      chunks: ["app"],
    }),
  ],
};
