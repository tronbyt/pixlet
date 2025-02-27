import { merge } from "webpack-merge";
import common from "./webpack.common.js";

import CopyWebpackPlugin from "copy-webpack-plugin";
import HtmlWebPackPlugin from "html-webpack-plugin";

const htmlPlugin = new HtmlWebPackPlugin({
    template: "./src/index.html",
    filename: "./index.html",
    favicon: "src/favicon.png",
});

const copyPlugin = new CopyWebpackPlugin({
    patterns: [{ from: "public", to: "../" }],
});

let plugins = [htmlPlugin, copyPlugin];

export default merge(common, {
    mode: "development",
    devtool: "source-map",
    devServer: {
        port: 3000,
        historyApiFallback: true,
        proxy: [
            {
                context: ["/api"],
                target: "http://localhost:8080",
                ws: true,
            },
        ],
    },
    plugins: plugins,
});
