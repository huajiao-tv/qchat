/* eslint-disable */
/** RC4.decode(data:String/Uint8Array, key:String):String/Uint8Array
 * @description     given a data string encoded with the same key
 *                  generates original data string.
 * @param   String  data encoded using same key
 * @param   String  key precedently used to encode data
 * @return  String  decoded data
 */
function decode(data, key) {
    return encode(data, key);
}

/** RC4.encode(data:String/Uint8Array, key:String):String/Uint8Array
 * @description     encode a data string using provided key
 * @param   String  key to use for this encoding
 * @param   String  data to encode
 * @return  String  encoded data. Will require same key to be decoded
 */
function encode(data, key) {
    var s = [],
        i, j = 0,
        x, y,
        res = [],
        u8arr, len;

    if(data instanceof ArrayBuffer) {
        data = new Uint8Array(data);
    }
    u8arr = data instanceof Uint8Array;
    len = u8arr ? data.byteLength : data.length;

    for(i = 0; i < 256; i++) {
        s[i] = i;
    }
    for(i = 0; i < 256; i++) {
        j = (j + s[i] + key.charCodeAt(i % key.length)) % 256;
        x = s[i];
        s[i] = s[j];
        s[j] = x;
    }
    i = 0;
    j = 0;
    for(y = 0; y < len; y++) {
        i = (i + 1) % 256;
        j = (j + s[i]) % 256;
        x = s[i];
        s[i] = s[j];
        s[j] = x;
        res[y] = (u8arr ? data[y] : data.charCodeAt(y)) ^ s[(s[i] + s[j]) % 256];
        if(!u8arr) {
            res[y] = String.fromCharCode(res[y]);
        }
    }
    return u8arr ? new Uint8Array(res) : res.join('');
}

export default {
    decode,
    encode
};
