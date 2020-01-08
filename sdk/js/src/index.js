import EventEmitter from 'wolfy87-eventemitter';
import ByteBuffer from 'bytebuffer';
import pako from 'pako';

import Logger from './utils/Logger'
import StringUtil from './utils/String'
import SecurityUtil from './utils/Security'
import NumberUtil from './utils/Number'
import projectConfig from '../package'
import Socket from './websocket/index';
import protobuf from "./proto/bundle";

import CONSTANTS from './constants';

import RC4 from './utils/RC4';
import Timer from './utils/Timer';
import md5 from "./utils/MD5";

export default class LiveSocket {

    // 配置
    config = {
        // 必传
        flag: '',
        protocolVersion: 0,
        clientVersion: 0,
        appId: 0,
        reserved: '',
        defaultKey: '',
        senderType: '',
        wsServer: '',

        // 非必传
        uid: '',
        sign: '',

        heartbeatInterval: 5000,
        sender: '', // 用户标识, userId 或 大于等于 20 位的纯数字
        password: '', // 加解密标识, = sender
    };

    // 状态记录
    state = {
        connectTryTimes: 0, // WebSocket 连接失败重连计数
        connected: false, // 是否已连接
        handshake: false, // 是否与服务端握手成功
        login: false,     // 是否登录成功

        roomId: '', // 当前在线房间 Id
        idSeed: '', // 房间操作时候用以生成唯一操作 Id

        sn: 0, // 消息sn, 消息发送时会生成一个sn, 在响应时会返回对应sn, 可以用来响应指定请求
        session: '', // 加解密用
    };

    // 事件对象
    event;

    // 定时器管理器
    timer;

    // WebSocket 实例
    socket;

    /**
     */
    constructor(config) {
        Logger.debug(`version: ${projectConfig.version}`);
        // 验证必传属性 是否传递 及 类型是否匹配
        [
            'flag',
            'protocolVersion',
            'clientVersion',
            'appId',
            'reserved',
            'defaultKey',
            'senderType',
            'wsServer',
        ].forEach((property) => {
            // 验证是否传入 及 类型是否正确
            let type = typeof config[property];
            if(type === 'undefined') {
                const errMsg = `${property} 不能为空`;
                Logger.error(errMsg);
                throw new Error(errMsg);
            }
            let targetType = typeof this.config[property];
            if(type !== targetType) {
                const errMsg = `${property} 类型要求为 ${targetType}`;
                Logger.error(errMsg);
                throw new Error(errMsg);
            }
        });
        // 验证非必传属性 若有传递, 其类型是否匹配
        [
            'heartbeatInterval',
            'uid',
            'sign',
        ].forEach((property) => {
            let type = typeof config[property];
            if(type !== 'undefined') {
                // 类型是否正确
                let targetType = typeof this.config[property];
                if(type !== targetType) {
                    const errMsg = `${property} 类型要求为 ${targetType}`;
                    Logger.error(errMsg);
                    throw new Error(errMsg);
                }
            }
        });
        // 特殊验证
        if((config.uid && !config.sign) || (config.sign && !config.uid)) {
            let errMsg = `需要登录用户时, uid sign 都必须传入, 当前配置 ( uid = ${config.uid}, sign = ${config.sign} ) `;
            Logger.debug(errMsg);
            throw new Error(errMsg);
        }
        // 覆盖配置
        Object.assign(this.config, config);

        // 处理特殊配置
        if(this.config.uid) {
            this.config.sender = this.config.uid;
            Logger.debug(`登录用户ID: ${this.config.sender}`);
        } else {
            this.config.sender = StringUtil.random(11, true) + Date.now();
            Logger.debug(`游客ID: ${this.config.sender}`);
        }
        this.config.password = this.config.sender;

        // 事件对象
        this.event = new EventEmitter();

        // 定时器管理器
        this.timer = new Timer();
    }

    connect() {
        return new Promise(resolve => {
            this.event.on('InitSuccess', resolve);
            Logger.debug('connect.');
            Socket.connectSocket({
                url: this.config.wsServer,
                binaryType: 'arraybuffer',
            }).then((task) => {
                this.socket = task;
                this.socket.onOpen(this.handleWebSocketOpen.bind(this));
                this.socket.onMessage(this.handleWebSocketMessage.bind(this));
                this.socket.onClose(this.handleWebSocketClose.bind(this));
                this.socket.onError(this.handleWebSocketError.bind(this));
            });
        })
    }

    // 断开
    close() {
        Logger.debug('close.');
        this.sendQuitRoomPack();
        this.timer.removeAll();
        this.socket.close();
        this.socket = null;
        this.state = {
            connectTryTimes: 0, // WebSocket 连接失败重连计数
            connected: false, // 是否已连接
            handshake: false, // 是否与服务端握手成功
            login: false,     // 是否登录成功

            roomId: '', // 当前在线房间 Id
            idSeed: '', // 房间操作时候用以生成唯一操作 Id

            sn: 0, // 消息sn, 消息发送时会生成一个sn, 在响应时会返回对应sn, 可以用来响应指定请求
            session: '', // 加解密用
        };
    }

    // 重连
    reconnect() {
        Logger.debug('reconnect.');
        this.close();
        this.connect();
    }

    handleWebSocketOpen() {
        this.state.connectTryTimes = 0;
        this.state.connected = true;
        if(!this.state.handshake) {
            this.sendHandShakePack();
        }
    }

    handleWebSocketMessage(event) {
        if(!this.state.handshake || !this.state.login) {
            try {
                if(!this.state.handshake) {
                    // 第一个包响应握手包
                    this.processHandShakePack(event.data);
                } else if(!this.state.login) {
                    this.processLoginPack(event.data);
                }
            } catch (e) {
                this.event.removeListener('InitSuccess')
            }
        } else {
            this.processMessagePack(event.data);
        }
    }

    handleWebSocketClose(e) {
        // 1000 为 正常关闭
        if(e.code !== 1000) {
            Logger.error('WebSocket异常关闭，尝试重连', e);
            let num = 1;
            this.timer.interval('err_close', () => {
                if(!this.state.connected) {
                    Logger.warn(`直播间重连第${num}次...`);
                    num += 1;
                    this.reconnect();
                } else {
                    this.timer.removeInterval(err_close);
                }
            }, 2000);
        }
    }

    handleWebSocketError(event) {
        Logger.error('WebSocket error observed:', event);
    }

    send(msg) {
        if(this.state.connected && this.socket) {
            this.socket.send({ data: msg });
        }
    }

    sendHandShakePack() {
        // magic + protobuf
        const bb = new ByteBuffer(12);
        // write magic = flag(2bytes) + protocol_version(1bytes) + client_version(1bytes) + appid(2bytes) + reserved(6bytes)
        // 12 为 magic 的长度
        bb
        // flag(2bytes)
            .writeString(this.config.flag)
            // protocol_version(1bytes)
            .writeInt8(this.config.protocolVersion << 4)
            // client_version(1bytes)
            .writeInt8(this.config.clientVersion)
            // appid(2bytes)
            .writeInt16(this.config.appId)
            // reserved(6bytes)
            .writeInt32(0)
            .writeInt16(0);

        let len = 0;
        // length of magic
        len += bb.view.byteLength;

        // 生成业务 protobuf 对象
        const msg = this.newMessagesMessageRequest(CONSTANTS.MESSAGE_ID.InitLoginReq, {
            init_login_req: this.newProtoMessage('qihoo.protocol.messages.InitLoginReq', {
                client_ram: StringUtil.random(10),
                sig: this.config.sign,
            }),
        });

        // RC4 加密 buffer
        const encryptMsg = RC4.encode(msg.toArrayBuffer(), this.config.defaultKey);

        // length of len
        len += 4;
        // length of encryptMsg
        len += encryptMsg.byteLength;
        bb
            .writeInt32(len)
            .append(encryptMsg)
            .flip();
        // send Message
        this.send(bb.toArrayBuffer());
    }

    /**
     * 响应握手包
     * @param data flag(2bytes) + len(4bytes) + Message(protobuf)
     */
    processHandShakePack(data) {
        const bb = new ByteBuffer();
        bb.append(data);

        // 验证 flag
        const flag = bb.readString(2, ByteBuffer.METRICS_CHARS, 0).string;
        if(flag !== this.config.flag) {
            let errMsg = '服务器响应标识 ( flag ) 有误';
            Logger.error(errMsg);
            throw new Error(errMsg);
        }

        // const length = bb.readInt32(2); // 包长度

        // 解析 Protobuf
        const msgBuffer = bb.slice(6, bb.view.byteLength);
        let msg;
        try {
            msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.config.defaultKey).buffer);
        } catch (e) {
            let errMsg = '解析 HandShake 消息体异常';
            Logger.error(errMsg, e);
            throw new Error(errMsg);
        }
        // 验证 msgid
        if(msg.msgid !== CONSTANTS.MESSAGE_ID.InitLoginResp) {
            let errMsg = `HandShake MsgId 验证失败, msgId: ${msg.msgid}`;
            Logger.error(errMsg);
            throw new Error(errMsg);
        }
        // 验证 sn
        if(msg.sn !== this.state.sn) {
            let errMsg = `HandShake SN 验证失败, msgId: ${msg.msgid}`;
            Logger.error(errMsg);
            throw new Error(errMsg);
        }
        this.state.handshake = true;

        this.event.emit('HandShake');
        Logger.debug('handshake success.');
        this.sendLoginPack(msg.resp.init_login_resp.server_ram);
    }

    sendLoginPack(serverRam) {
        const bb = new ByteBuffer();
        let len = 0;

        const msg = this.newMessagesMessageRequest(CONSTANTS.MESSAGE_ID.LoginReq, {
            login: this.newProtoMessage('qihoo.protocol.messages.LoginReq', {
                app_id: this.config.appId,
                server_ram: serverRam,
                // secret_ram = RC4(serverRam + randomString(8))
                secret_ram: RC4.encode(
                    new ByteBuffer(serverRam.length + 8)
                        .writeString(serverRam + StringUtil.random(8))
                        .view,
                    this.config.password
                ),
                verf_code: SecurityUtil.makeVerfCode(this.config.sender),
                // net_type = 0:unkonwn 1:2g 2:3g 3:wifi 4:ethe 5:4g
                // fixme 能否通过方法找到网络环境
                net_type: 4,
                // android, ios, pc
                mobile_type: 'pc',
                not_encrypt: true,
                platform: 'web',
            })
        });

        const encryptMsg = RC4.encode(msg.toArrayBuffer(), this.config.defaultKey);

        len += encryptMsg.byteLength;
        len += 4;

        bb
            .writeInt32(len)
            .append(encryptMsg)
            .flip();

        Logger.debug('send LoginPackage.');
        this.send(bb.toArrayBuffer());
    }

    /**
     * 响应登录包
     * @param data len(4bytes) + Message(protobuf)
     */
    processLoginPack(data) {
        const bb = new ByteBuffer();
        bb.append(data);

        const msgBuffer = bb.slice(4, bb.view.byteLength);
        let msg;
        try {
            if(this.state.login) {
                if(this.state.session) {
                    msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.state.session));
                } else {
                    msg = this.parseMessagesMessage(msgBuffer.toArrayBuffer());
                }
            } else if(this.config.password) {
                msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.config.password));
            }
        } catch (e) {
            msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.config.defaultKey));
        }
        if(msg.msgid !== CONSTANTS.MESSAGE_ID.LoginResp) {
            let errMsg = `Login MsgId 验证失败, msgId: ${msg.msgid}`;
            Logger.error(errMsg);
            throw new Error(errMsg);
        }
        // 验证 sn
        if(msg.sn !== this.state.sn) {
            let errMsg = `Login SN 验证失败, msgId: ${msg.msgid}`;
            Logger.error(errMsg);
            throw new Error(errMsg);
        }
        if(msg.resp.error) {
            Logger.error(msg.resp.error, StringUtil.Uint8ArrayToString(msg.resp.error.description));
            throw new Error(msg.resp.error)
        }

        this.state.login = true;
        this.state.session = msg.resp.login.session_key;

        this.event.emit('Login');
        this.event.emit('InitSuccess');
        Logger.debug('Login Success.');

        this.initHeartBeat();
    }

    joinRoom(roomId) {
        return new Promise((resolve, reject) => {
            this.event.addOnceListener('JoinRoomSuccess', () => {
                resolve();
                this.event.removeListener('JoinRoomFail', reject);
            });
            this.event.addOnceListener('JoinRoomFail', () => {
                reject();
                this.event.removeListener('JoinRoomSuccess', resolve);
            });
            if(!roomId) {
                Logger.error('roomId 不能为空');
                return;
            }
            if(typeof roomId !== 'string') {
                Logger.error('roomId 必须为 string 类型');
                return;
            }
            const bb = new ByteBuffer();
            let len = 0;

            const roomIdByte = new ByteBuffer(roomId.length).writeString(roomId).view;

            const chatroomChatRoomPacket = this.newChatroomChatRoomPacket({
                payloadtype: 102,
                applyjoinchatroomreq: this.newProtoMessage('qihoo.protocol.chatroom.ApplyJoinChatRoomRequest', {
                    roomid: roomIdByte,
                    room: this.newProtoMessage('qihoo.protocol.chatroom.ChatRoom', {
                        roomid: roomIdByte
                    }),
                    userid_type: 0
                })
            }, roomId);
            const msg = this.newMessagesMessageRequest(CONSTANTS.MESSAGE_ID.Service_Req, {
                service_req: this.newProtoMessage('qihoo.protocol.messages.Service_Req', {
                    service_id: 10000006,
                    request: chatroomChatRoomPacket.toArrayBuffer(),
                })
            });

            if(this.state.session && this.state.session !== '') {
                const encryptMsg = RC4.encode(msg.toArrayBuffer(), this.state.session);
                len += encryptMsg.byteLength;
                len += 4;
                bb
                    .writeInt32(len)
                    .append(encryptMsg)
                    .flip();
            } else {
                let msgBuffer = msg.toArrayBuffer();
                len += msgBuffer.byteLength;
                len += 4;
                bb
                    .writeInt32(len)
                    .append(msgBuffer)
                    .flip();
            }
            Logger.debug('send JoinRoom.');
            this.timer.timeout('JoinRoomTimeout', () => {
                this.event.emit('JoinRoomFail', 'timeout');
            }, 1000);
            this.send(bb.toArrayBuffer());
        })
    }

    quitRoom() {
        this.sendQuitRoomPack();
    }

    sendQuitRoomPack() {
        const bb = new ByteBuffer();
        let len = 0;

        const roomIdByte = new ByteBuffer(this.state.roomId.length).writeString(this.state.roomId).view;

        const chatroomChatRoomPacket = this.newChatroomChatRoomPacket({
            payloadtype: 103,
            quitchatroomreq: this.newProtoMessage('qihoo.protocol.chatroom.QuitChatRoomRequest', {
                roomid: roomIdByte,
                room: this.newProtoMessage('qihoo.protocol.chatroom.ChatRoom', {
                    roomid: roomIdByte
                })
            })
        }, this.state.roomId);
        const msg = this.newMessagesMessageRequest(CONSTANTS.MESSAGE_ID.Service_Req, {
            service_req: this.newProtoMessage('qihoo.protocol.messages.Service_Req', {
                service_id: 10000006,
                request: chatroomChatRoomPacket.toArrayBuffer(),
            })
        });

        if(this.state.session && this.state.session !== '') {
            const encryptMsg = RC4.encode(msg.toArrayBuffer(), this.state.session);
            len += encryptMsg.byteLength;
            len += 4;
            bb
                .writeInt32(len)
                .append(encryptMsg)
                .flip();
        } else {
            let msgBuffer = msg.toArrayBuffer();
            len += msgBuffer.byteLength;
            len += 4;
            bb
                .writeInt32(len)
                .append(msgBuffer)
                .flip();
        }
        Logger.debug('send QuitRoom.');
        this.timer.timeout('QuitRoomTimeout', () => {
            this.event.emit('QuitRoomFail', 'timeout');
        }, 1000);
        this.send(bb.toArrayBuffer());
    }

    sendPullMessagePack(infoType, infoId) {
        const bb = new ByteBuffer();
        let len = 0;

        const msg = this.newMessagesMessageRequest(CONSTANTS.MESSAGE_ID.GetInfoReq, {
            get_info: this.newProtoMessage('qihoo.protocol.messages.GetInfoReq', {
                info_type: infoType,
                get_info_id: infoId,
                get_info_offset: 1,
            }),
        });

        if(this.state.session && this.state.session !== '') {
            const encryptMsg = RC4.encode(msg.toArrayBuffer(), this.state.session);
            len += encryptMsg.byteLength;
            len += 4;
            bb
                .writeInt32(len)
                .append(encryptMsg)
                .flip();
        } else {
            let msgBuffer = msg.toArrayBuffer();
            len += msgBuffer.byteLength;
            len += 4;
            bb
                .writeInt32(len)
                .append(msgBuffer)
                .flip();
        }
        Logger.debug(`send PullMessagePack.(${infoType}, ${infoId})`);
        this.send(bb.toArrayBuffer());
    }

    // ================= 心跳相关 start =================

    initHeartBeat() {
        this.timer.interval('heartBeat', () => {
            this.sendHeartbeatPack();
        }, this.config.heartbeatInterval)
    }

    sendHeartbeatPack() {
        this.send(new ByteBuffer(4).writeInt32(0).buffer);
        Logger.debug('send Heartbeat.');
        this.timer.timeout('heartBeatTimeout', () => {
            this.timer.removeInterval('heartBeat');
            this.reconnect();
        }, 3000)
    }

    processHeartbeatPack() {
        this.timer.removeTimeout('heartBeatTimeout');
        Logger.debug('process Heartbeat.')
    }

    // ================= 心跳相关 end =================

    /**
     * 处理正常流程
     */
    processMessagePack(data) {
        const bb = new ByteBuffer();
        bb.append(data);

        // 处理心跳包
        if(
            bb.view.byteLength === 16
            && bb.readInt32(0) === 0
            && bb.readInt32(4) === 0
            && bb.readInt32(8) === 0
            && bb.readInt32(12) === 0) {
            this.processHeartbeatPack();
            return;
        }

        const msgBuffer = bb.slice(4, bb.view.byteLength);
        let msg;
        if(this.state.login) {
            if(this.state.session) {
                msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.state.session));
            } else {
                msg = this.parseMessagesMessage(msgBuffer.toArrayBuffer());
            }
        } else if(this.config.password) {
            msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.config.password));
        } else {
            msg = this.parseMessagesMessage(RC4.decode(msgBuffer.toArrayBuffer(), this.config.defaultKey));
        }

        if(msg.resp && msg.resp.error) {
            Logger.error(msg.resp.error, StringUtil.Uint8ArrayToString(msg.resp.error.description));
            throw new Error(msg.resp.error)
        }

        if(msg.msgid === 200011) {
            this.processServiceRespPack(msg);
        } else if(msg.msgid === 300000) {
            this.processNewMessageNotifyPack(msg);
        } else if(msg.msgid === 200004) {
            this.processGetInfoRespPack(msg);
        } else {
            Logger.debug(msg);
            this.event.emit('UnsupportedMsg', msg)
        }
        // GetInfoResp       200004
        // NewMessageNotify  300000
        // Service_Resp      200011
    }

    processServiceRespPack(msg) {
        // 验证 sn
        if(msg.sn !== this.state.sn) {
            let errMsg = `ServiceResp SN 验证失败, msgId: ${msg.msgid}`;
            Logger.error(errMsg);
            throw new Error(errMsg);
        }
        const message = this.parseChatroomChatRoomPacket(msg.resp.service_resp.response);
        const toUserData = message.to_user_data;

        if(toUserData.payloadtype === CONSTANTS.PAYLOADTYPE.applyjoinchatroomresp) {
            this.timer.removeTimeout('JoinRoomTimeout');
            if(toUserData.result === CONSTANTS.PAYLOADTYPE.successful) {
                this.state.roomId = StringUtil.Uint8ArrayToString(toUserData.applyjoinchatroomresp.room.roomid);
                this.event.emit('JoinRoomSuccess', toUserData.applyjoinchatroomresp);
                Logger.debug('JoinRoom success.');
            } else {
                // 踢人逻辑
                this.event.emit('JoinRoomFail', toUserData.applyjoinchatroomresp);
                Logger.debug('JoinRoom fail.');
            }
        } else if(toUserData.payloadtype === CONSTANTS.PAYLOADTYPE.quitchatroomresp) {
            this.timer.removeTimeout('QuitRoomTimeout');
            if(toUserData.result === CONSTANTS.PAYLOADTYPE.successful) {
                this.state.roomId = '';
                this.event.emit('QuitRoomSuccess', toUserData.quitchatroomresp);
                Logger.debug('QuitRoom success.');
            } else {
                this.event.emit('QuitRoomFail', toUserData.quitchatroomresp);
                Logger.debug('QuitRoom fail.');
            }
        }
    }

    processNewMessageNotifyPack(msg) {
        const infoType = msg.notify.newinfo_ntf.info_type;
        if(infoType === 'peer' || infoType === 'im') {
            const infoId = msg.notify.newinfo_ntf.info_id;
            // peer 消息
            this.sendPullMessagePack(infoType, infoId);
        } else {
            const message = this.parseChatroomChatRoomPacket(msg.notify.newinfo_ntf.info_content);
            let toUserData = message.to_user_data;
            if(toUserData && toUserData.result === CONSTANTS.PAYLOADTYPE.successful) {
                // 根据类型区分处理
                if(toUserData.payloadtype === CONSTANTS.PAYLOADTYPE.newmsgnotify) {
                    // msgtype = text, voice, link
                    if(toUserData.newmsgnotify.msgtype === 0) {
                        // text
                        let onlineCount = toUserData.newmsgnotify.memcount | 0;
                        let content;
                        if(toUserData.newmsgnotify.msgcontent) {
                            content = JSON.parse(StringUtil.Uint8ArrayToString(toUserData.newmsgnotify.msgcontent));
                        }
                        if(content) {
                            // console.log(onlineCount, JSON.stringify(content));
                            this.event.emit('MemberCountUpdate', onlineCount);
                            this.event.emit('Message', content);
                        }
                    }
                } else if(toUserData.payloadtype === CONSTANTS.PAYLOADTYPE.membergzipnotify) {
                    /*
                     * 压缩类型解析
                     */
                    try {
                        for(let i = 0; i < data.multinotify.length; i++) {
                            let buffer = pako.ungzip(
                                data.multinotify[i].data.buffer.slice(data.multinotify[i].data.offset,
                                    data.multinotify[i].data.limit)
                            );
                            let content = this.parseChatroomChatRoomNewMsg(buffer);
                            if(content) {
                                content = StringUtil.Uint8ArrayToString(content.msgcontent);
                                content = JSON.parse(content);
                                this.event.emit('Message', content);
                            }
                        }
                    } catch (e) {
                        Logger.error('压缩消息解析失败', e)
                    }
                }
            }
        }
    }

    processGetInfoRespPack(msg) {
        // {
        //     "msgid": 200004,
        //     "sn": "7485957344",
        //     "resp": {
        //         "get_info": {
        //              "info_type": "peer",
        //              "infos": [
        //                  {
        //                      "property_pairs": [
        //                          {
        //                              "key": "aW5mb19pZA==",
        //                              "value": "AAAAAAAAIZM="
        //                          },
        //                      ]
        //                  }
        //              ],
        //              "last_info_id": "8595"
        //          }
        //     }
        // }
        const infos = msg.resp.get_info.infos;
        infos.forEach((info) => {
            const propertyPairs = info.property_pairs;
            propertyPairs.forEach((pair) => {
                if(StringUtil.Uint8ArrayToString(pair.key) === 'chat_body') {
                    let p2MessageStr = StringUtil.Uint8ArrayToString(pair.value);
                    Logger.debug('processGetInfoRespMessage:', p2MessageStr);
                    try {
                        p2MessageStr = JSON.parse(p2MessageStr);
                    } catch (e) {
                    }
                    this.event.emit('PeerMessage', p2MessageStr);
                }
            });
        });

    }

    // ================= GoogleProtobuf 相关处理 start =================
    /**
     * 获取指定 Proto 原型
     * @param protoPackage Proto 路径
     * @return Proto 原型
     */
    getProtoClass(protoPackage) {
        const props = protoPackage.split('.');
        let entry = protobuf;
        props.forEach((v) => {
            if(entry[v]) {
                entry = entry[v];
            } else {
                throw new Error('not fund');
            }
        });
        return entry;
    }

    /**
     * 生成 Proto 对象
     * @param classPath 类路径
     * @param params    构造参数
     * @return Proto 对象实例
     */
    newProtoMessage(classPath, params) {
        const Message = this.getProtoClass(classPath);
        const entry = Message.create(params);
        entry.toArrayBuffer = function () { // eslint-disable-line
            return new Uint8Array(Message.encode(entry).finish());
        };
        return entry;
    }

    /**
     * 解析 proto 包
     */
    parseProto(classPath, data) {
        return this.getProtoClass(classPath).decode(data);
    }

    // 流程包解析
    parseMessagesMessage(data) {
        // 这里传入的都是new出来的ByteBuffer。统一转为Unit8Array
        return this.parseProto('qihoo.protocol.messages.Message', new Uint8Array(data));
    }

    // 普通消息解析
    parseChatroomChatRoomPacket(data) {
        return this.parseProto('qihoo.protocol.chatroom.ChatRoomPacket', data);
    }

    // 压缩消息解析
    parseChatroomChatRoomNewMsg(data) {
        return this.parseProto('qihoo.protocol.chatroom.ChatRoomNewMsg', data);
    }

    // ================= GoogleProtobuf 相关处理 end =================

    /**
     * 通用请求对象
     * @param msgId
     * @param request
     * @private
     */
    newMessagesMessageRequest(msgId, request) {
        this.state.sn = NumberUtil.random(10);
        return this.newProtoMessage('qihoo.protocol.messages.Message', {
            msgid: msgId,
            sn: this.state.sn,
            sender: this.config.sender,
            sender_type: this.config.senderType,
            req: this.newProtoMessage('qihoo.protocol.messages.Request', request)
        });
    }

    /**
     * 房间操作请求对象
     * @param toServerData
     * @param rid 房间 Id
     */
    newChatroomChatRoomPacket(toServerData, rid) {
        if(!rid) {
            rid = this.state.roomId
        }
        const roomId = new ByteBuffer(rid.length).writeString(rid).view;

        return this.newProtoMessage('qihoo.protocol.chatroom.ChatRoomPacket', {
            client_sn: this.state.sn,
            roomid: roomId,
            appid: this.config.appId,
            uuid: md5(
                StringUtil.random(10) +
                StringUtil.leftPad(++this.state.idSeed, 10, '0') +
                Date.now()
            ),
            to_server_data: this.newProtoMessage('qihoo.protocol.chatroom.ChatRoomUpToServer', toServerData)
        });
    }
}
