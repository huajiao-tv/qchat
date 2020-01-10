
db = db.getSiblingDB("msg_im_inboxes");
db.createCollection("msg_im_inbox_1090");
db.msg_im_inbox_1090.createIndex({"owner":"hashed"});
db.msg_im_inbox_1090.createIndex({"owner":1,"msg_id":1},{"unique":true});
db.msg_im_inbox_1090.createIndex({"count_time":1},{"expireAfterSeconds":604800});
db.msg_im_inbox_1090.createIndex({"msg_ctime":1});
db.msg_im_inbox_1090.createIndex({"modified":1});

db = db.getSiblingDB("msg_im_lastr");
db.createCollection("im_last_reads_1090");
db.im_last_reads_1090.createIndex({"jid":"hashed"});
db.im_last_reads_1090.createIndex({"jid":1},{"unique":true});
db.im_last_reads_1090.createIndex({"latest_modified":1});

db = db.getSiblingDB("msg_im_outboxes");
db.createCollection("msg_im_outbox_1090");
db.msg_im_outbox_1090.createIndex({"owner":"hashed"});
db.msg_im_outbox_1090.createIndex({"owner":1,"msg_id":1},{"unique":true});
db.msg_im_outbox_1090.createIndex({"count_time":1},{"expireAfterSeconds":604800});
db.msg_im_outbox_1090.createIndex({"msg_ctime":1});
db.msg_im_outbox_1090.createIndex({"modified":1});

db = db.getSiblingDB("msg_imid");
db.createCollection("im_latest_id_1090");
db.im_latest_id_1090.createIndex({"jid":"hashed"});
db.im_latest_id_1090.createIndex({"jid":1},{"unique":true});
db.im_latest_id_1090.createIndex({"latest_modified":1});

db = db.getSiblingDB("msg_lastr");
db.createCollection("last_reads_1090");
db.last_reads_1090.createIndex({"jid":"hashed"});
db.last_reads_1090.createIndex({"jid":1},{"unique":true});
db.last_reads_1090.createIndex({"latest_modified":1});

db = db.getSiblingDB("msg_messageid");
db.createCollection("latest_id_1090");
db.latest_id_1090.createIndex({"jid":"hashed"});
db.latest_id_1090.createIndex({"jid":1},{"unique":true});
db.latest_id_1090.createIndex({"latest_modified":1});

db = db.getSiblingDB("msg_messages");
db.createCollection("message_data_1090");
db.message_data_1090.createIndex({"jid":"hashed"});
db.message_data_1090.createIndex({"jid":1,"msg_id":1},{"unique":true});
db.message_data_1090.createIndex({"count_time":1},{"expireAfterSeconds":604800});
db.message_data_1090.createIndex({"msg_ctime":1});
db.message_data_1090.createIndex({"modified":1});

db = db.getSiblingDB("msg_publics");
db.createCollection("public_message_data_1090");
db.public_message_data_1090.createIndex({"msg_id":"hashed"});
db.public_message_data_1090.createIndex({"msg_id":1},{"unique":true});
db.public_message_data_1090.createIndex({"count_time":1},{"expireAfterSeconds":604800});
