export default class StringUtil {
    static random(len, onlyNumber = false) {
        const str = [];
        const seedChars = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
        const arr = onlyNumber ? seedChars.slice(0, 10) : seedChars;
        for(let i = 0; i < len; i++) {
            str.push(arr.charAt(Math.floor(Math.random() * arr.length)));
        }
        return str.join('');
    }

    static leftPad(string, size, character) {
        let result = String(string);
        character = character || ' ';
        while(result.length < size) {
            result = character + result;
        }
        return result;
    }

    static Uint8ArrayToString(array) {
        let out, i, len, c;
        let char2, char3;

        out = "";
        len = array.length;
        i = 0;
        while(i < len) {
            c = array[i++];
            switch(c >> 4) {
                case 0:
                case 1:
                case 2:
                case 3:
                case 4:
                case 5:
                case 6:
                case 7:
                    // 0xxxxxxx
                    out += String.fromCharCode(c);
                    break;
                case 12:
                case 13:
                    // 110x xxxx 10xx xxxx
                    char2 = array[i++];
                    out += String.fromCharCode(((c & 0x1F) << 6) | (char2 & 0x3F));
                    break;
                case 14:
                    // 1110 xxxx 10xx xxxx 10xx xxxx
                    char2 = array[i++];
                    char3 = array[i++];
                    out += String.fromCharCode(((c & 0x0F) << 12) |
                        ((char2 & 0x3F) << 6) |
                        ((char3 & 0x3F) << 0));
                    break;
            }
        }
        return out;
    }
}
