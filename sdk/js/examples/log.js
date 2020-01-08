const element = document.querySelector('#log');
window.console = {
    wc: window.console
};    //将原本console的引用指向console的一个成员变量wc，以便在后面扩充的函数中使用。
['log', 'error', 'warn', 'debug', 'info'].forEach(function (item) {  //针对console的五种方法
    console[item] = function (...msg) {
        this.wc[item](...msg);
        let element1 = document.createElement('p');
        element1.innerHTML = `${msg}`;
        element1.classList.add(item);
        element.appendChild(element1)
    }
});