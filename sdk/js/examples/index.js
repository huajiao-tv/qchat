import LiveSocket from '../src/index';
// import LiveSocket from '../dist/index';
import './log';

let webSocketIO2;


document.querySelector('#connect').addEventListener('click', () => {
    let config = JSON.parse(document.querySelector('#config').value);
    webSocketIO2 = new LiveSocket(config);
    webSocketIO2.event.on('Message', (data) => {
        console.log(JSON.stringify(data));
    });
    webSocketIO2.connect()
        .then(() => {
            console.log("嘿嘿副科级暗红色的扣积分");
        });
});

document.querySelector('#join').addEventListener('click', () => {
    let config = document.querySelector('#roomId').value;
    webSocketIO2.joinRoom(config.trim())
        .then(() => {
            console.info('joinroom success');
        })
        .catch(() => {
            console.error('joinroom fail');
        })
});

document.querySelector('#quit').addEventListener('click', () => {
    webSocketIO2.quitRoom()
});
