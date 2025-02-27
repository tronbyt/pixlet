import { merge } from "webpack-merge";
import common from "./webpack.common.js";
import { resolve } from "path";
import HtmlWebPackPlugin from "html-webpack-plugin";
import CopyWebpackPlugin from "copy-webpack-plugin";

const htmlPlugin = new HtmlWebPackPlugin({
    template: "./src/index.html",
    filename: "../index.html",
    favicon: "src/favicon.png",
});

const copyPlugin = new CopyWebpackPlugin({
    patterns: [{ from: "public", to: "../" }],
});

let plugins = [htmlPlugin, copyPlugin];

export default merge(common, {
    mode: "production",
    devtool: false,
    output: {
        asyncChunks: true,
        publicPath: "auto",
        path: resolve(import.meta.dirname, "dist/static"),
        filename: "[name].[chunkhash].js",
        clean: true,
    },
    performance: {
        // free-brands-svg-icons and free-solid-svg-icons are large
        // libraries. They arey are bundled fully to look up arbitrary
        // icons by the name specified in the schema (see FieldIcon.jsx).
        // Increase the maxAssetSize to silence the warning.
        maxAssetSize: 1_000_000,
    },
    optimization: {
        // Creates a runtime file to be shared for all generated chunks.
        runtimeChunk: 'single'
    },
    plugins: plugins,
});
