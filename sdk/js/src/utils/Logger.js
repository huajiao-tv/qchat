export default class Logger {

    static prefix = '[ LiveSocket ]';

    static warn() {
        console.warn(this.prefix, ...arguments)
    }

    static info() {
        console.info(this.prefix, ...arguments)
    }

    static error() {
        console.error(this.prefix, ...arguments)
    }

    static debug() {
        console.debug(this.prefix, ...arguments)
    }
}
