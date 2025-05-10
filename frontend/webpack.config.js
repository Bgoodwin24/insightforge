const path = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

module.exports = {
  entry: "./src/index.jsx",
  output: {
    path: path.resolve(__dirname, "dist"),
    filename: "bundle.js",
    publicPath: "/",
  },
  mode: "development",
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/i,
        exclude: /node_modules/,
        use: "babel-loader",
      },
      {
        test: /\.module\.css$/, // For CSS Modules
        use: [
          MiniCssExtractPlugin.loader, // Extract CSS to a separate file
          {
            loader: "css-loader",
            options: {
              modules: {
                localIdentName: "[name]__[local]__[hash:base64:5]",
                exportLocalsConvention: "camelCase",
              },
            },
          },
        ],
      },
      {
        test: /\.css$/, // For Global CSS
        exclude: /\.module\.css$/,
        use: [
          MiniCssExtractPlugin.loader, // Extract global CSS
          "css-loader",
        ],
      },
      {
        test: /\.(png|jpg|jpeg|gif|svg)$/i,
        type: "asset/resource", // For images and static assets
      },
    ],
  },
  resolve: {
    extensions: [".js", ".jsx"], // Resolve file extensions
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: path.resolve(__dirname, "public", "index.html"), // Use the index.html template
      inject: true, // Automatically inject the generated JS and CSS files
    }),
    new MiniCssExtractPlugin({
      filename: "[name].[contenthash].css", // Customize CSS filename with content hash
    }),
  ],
  devServer: {
    static: {
      directory: path.join(__dirname, "public"),
    },
    historyApiFallback: true,
    port: 3000,
    open: true,
  },
};
