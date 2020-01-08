import md5 from './MD5';

export default class SecurityUtil {
    static makeVerfCode(jid) {
        const salt = '360tantan@1408$';
        return md5(jid + salt).substring(24);
    }
}
