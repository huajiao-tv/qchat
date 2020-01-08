import StringUtil from './String';

export default class NumberUtil {
    static random(len) {
        return parseInt(StringUtil.random(len, true), 10);
    }
}
