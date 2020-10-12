// source: m3env.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

goog.exportSymbol('proto.m3api.EnvListMsg', null, global);
goog.exportSymbol('proto.m3api.EnvMsg', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.m3api.EnvMsg = function (opt_data) {
    jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.m3api.EnvMsg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
    /**
     * @public
     * @override
     */
    proto.m3api.EnvMsg.displayName = 'proto.m3api.EnvMsg';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.m3api.EnvListMsg = function (opt_data) {
    jspb.Message.initialize(this, opt_data, 0, -1, proto.m3api.EnvListMsg.repeatedFields_, null);
};
goog.inherits(proto.m3api.EnvListMsg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
    /**
     * @public
     * @override
     */
    proto.m3api.EnvListMsg.displayName = 'proto.m3api.EnvListMsg';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
    /**
     * Creates an object representation of this proto.
     * Field names that are reserved in JavaScript and will be renamed to pb_name.
     * Optional fields that are not set will be set to undefined.
     * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
     * For the list of reserved names please see:
     *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
     * @param {boolean=} opt_includeInstance Deprecated. whether to include the
     *     JSPB instance for transitional soy proto support:
     *     http://goto/soy-param-migration
     * @return {!Object}
     */
    proto.m3api.EnvMsg.prototype.toObject = function (opt_includeInstance) {
        return proto.m3api.EnvMsg.toObject(opt_includeInstance, this);
    };


    /**
     * Static version of the {@see toObject} method.
     * @param {boolean|undefined} includeInstance Deprecated. Whether to include
     *     the JSPB instance for transitional soy proto support:
     *     http://goto/soy-param-migration
     * @param {!proto.m3api.EnvMsg} msg The msg instance to transform.
     * @return {!Object}
     * @suppress {unusedLocalVariables} f is only used for nested messages
     */
    proto.m3api.EnvMsg.toObject = function (includeInstance, msg) {
        var f, obj = {
            envId: jspb.Message.getFieldWithDefault(msg, 1, 0),
            schemaName: jspb.Message.getFieldWithDefault(msg, 2, ""),
            schemaSize: jspb.Message.getFieldWithDefault(msg, 3, 0),
            schemaSizePercent: jspb.Message.getFloatingPointFieldWithDefault(msg, 4, 0.0)
        };

        if (includeInstance) {
            obj.$jspbMessageInstance = msg;
        }
        return obj;
    };
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.m3api.EnvMsg}
 */
proto.m3api.EnvMsg.deserializeBinary = function (bytes) {
    var reader = new jspb.BinaryReader(bytes);
    var msg = new proto.m3api.EnvMsg;
    return proto.m3api.EnvMsg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.m3api.EnvMsg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.m3api.EnvMsg}
 */
proto.m3api.EnvMsg.deserializeBinaryFromReader = function (msg, reader) {
    while (reader.nextField()) {
        if (reader.isEndGroup()) {
            break;
        }
        var field = reader.getFieldNumber();
        switch (field) {
            case 1:
                var value = /** @type {number} */ (reader.readInt32());
                msg.setEnvId(value);
                break;
            case 2:
                var value = /** @type {string} */ (reader.readString());
                msg.setSchemaName(value);
                break;
            case 3:
                var value = /** @type {number} */ (reader.readInt64());
                msg.setSchemaSize(value);
                break;
            case 4:
                var value = /** @type {number} */ (reader.readFloat());
                msg.setSchemaSizePercent(value);
                break;
            default:
                reader.skipField();
                break;
        }
    }
    return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.m3api.EnvMsg.prototype.serializeBinary = function () {
    var writer = new jspb.BinaryWriter();
    proto.m3api.EnvMsg.serializeBinaryToWriter(this, writer);
    return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.m3api.EnvMsg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.EnvMsg.serializeBinaryToWriter = function (message, writer) {
    var f = undefined;
    f = message.getEnvId();
    if (f !== 0) {
        writer.writeInt32(
            1,
            f
        );
    }
    f = message.getSchemaName();
    if (f.length > 0) {
        writer.writeString(
            2,
            f
        );
    }
    f = message.getSchemaSize();
    if (f !== 0) {
        writer.writeInt64(
            3,
            f
        );
    }
    f = message.getSchemaSizePercent();
    if (f !== 0.0) {
        writer.writeFloat(
            4,
            f
        );
    }
};


/**
 * optional int32 env_id = 1;
 * @return {number}
 */
proto.m3api.EnvMsg.prototype.getEnvId = function () {
    return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.EnvMsg} returns this
 */
proto.m3api.EnvMsg.prototype.setEnvId = function (value) {
    return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string schema_name = 2;
 * @return {string}
 */
proto.m3api.EnvMsg.prototype.getSchemaName = function () {
    return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.m3api.EnvMsg} returns this
 */
proto.m3api.EnvMsg.prototype.setSchemaName = function (value) {
    return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional int64 schema_size = 3;
 * @return {number}
 */
proto.m3api.EnvMsg.prototype.getSchemaSize = function () {
    return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.EnvMsg} returns this
 */
proto.m3api.EnvMsg.prototype.setSchemaSize = function (value) {
    return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional float schema_size_percent = 4;
 * @return {number}
 */
proto.m3api.EnvMsg.prototype.getSchemaSizePercent = function () {
    return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 4, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.EnvMsg} returns this
 */
proto.m3api.EnvMsg.prototype.setSchemaSizePercent = function (value) {
    return jspb.Message.setProto3FloatField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.m3api.EnvListMsg.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
    /**
     * Creates an object representation of this proto.
     * Field names that are reserved in JavaScript and will be renamed to pb_name.
     * Optional fields that are not set will be set to undefined.
     * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
     * For the list of reserved names please see:
     *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
     * @param {boolean=} opt_includeInstance Deprecated. whether to include the
     *     JSPB instance for transitional soy proto support:
     *     http://goto/soy-param-migration
     * @return {!Object}
     */
    proto.m3api.EnvListMsg.prototype.toObject = function (opt_includeInstance) {
        return proto.m3api.EnvListMsg.toObject(opt_includeInstance, this);
    };


    /**
     * Static version of the {@see toObject} method.
     * @param {boolean|undefined} includeInstance Deprecated. Whether to include
     *     the JSPB instance for transitional soy proto support:
     *     http://goto/soy-param-migration
     * @param {!proto.m3api.EnvListMsg} msg The msg instance to transform.
     * @return {!Object}
     * @suppress {unusedLocalVariables} f is only used for nested messages
     */
    proto.m3api.EnvListMsg.toObject = function (includeInstance, msg) {
        var f, obj = {
            envsList: jspb.Message.toObjectList(msg.getEnvsList(),
                proto.m3api.EnvMsg.toObject, includeInstance)
        };

        if (includeInstance) {
            obj.$jspbMessageInstance = msg;
        }
        return obj;
    };
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.m3api.EnvListMsg}
 */
proto.m3api.EnvListMsg.deserializeBinary = function (bytes) {
    var reader = new jspb.BinaryReader(bytes);
    var msg = new proto.m3api.EnvListMsg;
    return proto.m3api.EnvListMsg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.m3api.EnvListMsg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.m3api.EnvListMsg}
 */
proto.m3api.EnvListMsg.deserializeBinaryFromReader = function (msg, reader) {
    while (reader.nextField()) {
        if (reader.isEndGroup()) {
            break;
        }
        var field = reader.getFieldNumber();
        switch (field) {
            case 1:
                var value = new proto.m3api.EnvMsg;
                reader.readMessage(value, proto.m3api.EnvMsg.deserializeBinaryFromReader);
                msg.addEnvs(value);
                break;
            default:
                reader.skipField();
                break;
        }
    }
    return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.m3api.EnvListMsg.prototype.serializeBinary = function () {
    var writer = new jspb.BinaryWriter();
    proto.m3api.EnvListMsg.serializeBinaryToWriter(this, writer);
    return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.m3api.EnvListMsg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.EnvListMsg.serializeBinaryToWriter = function (message, writer) {
    var f = undefined;
    f = message.getEnvsList();
    if (f.length > 0) {
        writer.writeRepeatedMessage(
            1,
            f,
            proto.m3api.EnvMsg.serializeBinaryToWriter
        );
    }
};


/**
 * repeated EnvMsg envs = 1;
 * @return {!Array<!proto.m3api.EnvMsg>}
 */
proto.m3api.EnvListMsg.prototype.getEnvsList = function () {
    return /** @type{!Array<!proto.m3api.EnvMsg>} */ (
        jspb.Message.getRepeatedWrapperField(this, proto.m3api.EnvMsg, 1));
};


/**
 * @param {!Array<!proto.m3api.EnvMsg>} value
 * @return {!proto.m3api.EnvListMsg} returns this
 */
proto.m3api.EnvListMsg.prototype.setEnvsList = function (value) {
    return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.m3api.EnvMsg=} opt_value
 * @param {number=} opt_index
 * @return {!proto.m3api.EnvMsg}
 */
proto.m3api.EnvListMsg.prototype.addEnvs = function (opt_value, opt_index) {
    return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.m3api.EnvMsg, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.m3api.EnvListMsg} returns this
 */
proto.m3api.EnvListMsg.prototype.clearEnvsList = function () {
    return this.setEnvsList([]);
};


goog.object.extend(exports, proto.m3api);
