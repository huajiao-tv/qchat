const UglifyJsPlugin = require('uglifyjs-webpack-plugin');

const polyfill = [];

const baseConfig = {
    entry: polyfill.concat(['./src/index.js']),
    devtool: false,
    module: {
        rules: [
            {
                test: /\.js$/,
                loader: 'babel-loader'
            }
        ]
    },
    mode: 'production',
    optimization: {
        minimizer: [new UglifyJsPlugin()],
    }
};

const umd = Object.assign({
    output: {
        path: `${__dirname}/dist`,
        filename: 'index.js',
        library: 'LiveSocket',
        libraryTarget: 'umd'
    },
}, baseConfig);

const client = Object.assign({
    output: {
        path: `${__dirname}/browser`,
        filename: 'index.js',
        library: 'LiveSocket',
        libraryTarget: 'window'
    },
}, baseConfig);

module.exports = [umd, client];
