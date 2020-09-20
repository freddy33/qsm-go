// source: m3path.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var m3point_pb = require('./m3point_pb.js');
goog.object.extend(proto, m3point_pb);
goog.exportSymbol('proto.m3api.NextMoveRespMsg', null, global);
goog.exportSymbol('proto.m3api.PathContextMsg', null, global);
goog.exportSymbol('proto.m3api.PathNodeMsg', null, global);
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
proto.m3api.PathContextMsg = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.m3api.PathContextMsg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.m3api.PathContextMsg.displayName = 'proto.m3api.PathContextMsg';
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
proto.m3api.PathNodeMsg = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.m3api.PathNodeMsg.repeatedFields_, null);
};
goog.inherits(proto.m3api.PathNodeMsg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.m3api.PathNodeMsg.displayName = 'proto.m3api.PathNodeMsg';
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
proto.m3api.NextMoveRespMsg = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.m3api.NextMoveRespMsg.repeatedFields_, null);
};
goog.inherits(proto.m3api.NextMoveRespMsg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.m3api.NextMoveRespMsg.displayName = 'proto.m3api.NextMoveRespMsg';
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
proto.m3api.PathContextMsg.prototype.toObject = function(opt_includeInstance) {
  return proto.m3api.PathContextMsg.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.m3api.PathContextMsg} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.PathContextMsg.toObject = function(includeInstance, msg) {
  var f, obj = {
    pathCtxId: jspb.Message.getFieldWithDefault(msg, 1, 0),
    growthContextId: jspb.Message.getFieldWithDefault(msg, 2, 0),
    growthOffset: jspb.Message.getFieldWithDefault(msg, 3, 0),
    center: (f = msg.getCenter()) && m3point_pb.PointMsg.toObject(includeInstance, f)
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
 * @return {!proto.m3api.PathContextMsg}
 */
proto.m3api.PathContextMsg.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.m3api.PathContextMsg;
  return proto.m3api.PathContextMsg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.m3api.PathContextMsg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.m3api.PathContextMsg}
 */
proto.m3api.PathContextMsg.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setPathCtxId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setGrowthContextId(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setGrowthOffset(value);
      break;
    case 4:
      var value = new m3point_pb.PointMsg;
      reader.readMessage(value,m3point_pb.PointMsg.deserializeBinaryFromReader);
      msg.setCenter(value);
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
proto.m3api.PathContextMsg.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.m3api.PathContextMsg.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.m3api.PathContextMsg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.PathContextMsg.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPathCtxId();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getGrowthContextId();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getGrowthOffset();
  if (f !== 0) {
    writer.writeInt32(
      3,
      f
    );
  }
  f = message.getCenter();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      m3point_pb.PointMsg.serializeBinaryToWriter
    );
  }
};


/**
 * optional int32 path_ctx_id = 1;
 * @return {number}
 */
proto.m3api.PathContextMsg.prototype.getPathCtxId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathContextMsg} returns this
 */
proto.m3api.PathContextMsg.prototype.setPathCtxId = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int32 growth_context_id = 2;
 * @return {number}
 */
proto.m3api.PathContextMsg.prototype.getGrowthContextId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathContextMsg} returns this
 */
proto.m3api.PathContextMsg.prototype.setGrowthContextId = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int32 growth_offset = 3;
 * @return {number}
 */
proto.m3api.PathContextMsg.prototype.getGrowthOffset = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathContextMsg} returns this
 */
proto.m3api.PathContextMsg.prototype.setGrowthOffset = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional PointMsg center = 4;
 * @return {?proto.m3api.PointMsg}
 */
proto.m3api.PathContextMsg.prototype.getCenter = function() {
  return /** @type{?proto.m3api.PointMsg} */ (
    jspb.Message.getWrapperField(this, m3point_pb.PointMsg, 4));
};


/**
 * @param {?proto.m3api.PointMsg|undefined} value
 * @return {!proto.m3api.PathContextMsg} returns this
*/
proto.m3api.PathContextMsg.prototype.setCenter = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.m3api.PathContextMsg} returns this
 */
proto.m3api.PathContextMsg.prototype.clearCenter = function() {
  return this.setCenter(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.m3api.PathContextMsg.prototype.hasCenter = function() {
  return jspb.Message.getField(this, 4) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.m3api.PathNodeMsg.repeatedFields_ = [7];



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
proto.m3api.PathNodeMsg.prototype.toObject = function(opt_includeInstance) {
  return proto.m3api.PathNodeMsg.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.m3api.PathNodeMsg} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.PathNodeMsg.toObject = function(includeInstance, msg) {
  var f, obj = {
    pathNodeId: jspb.Message.getFieldWithDefault(msg, 1, 0),
    pathCtxId: jspb.Message.getFieldWithDefault(msg, 2, 0),
    point: (f = msg.getPoint()) && m3point_pb.PointMsg.toObject(includeInstance, f),
    d: jspb.Message.getFieldWithDefault(msg, 4, 0),
    trioId: jspb.Message.getFieldWithDefault(msg, 5, 0),
    connectionMask: jspb.Message.getFieldWithDefault(msg, 6, 0),
    linkedPathNodeIdsList: (f = jspb.Message.getRepeatedField(msg, 7)) == null ? undefined : f
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
 * @return {!proto.m3api.PathNodeMsg}
 */
proto.m3api.PathNodeMsg.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.m3api.PathNodeMsg;
  return proto.m3api.PathNodeMsg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.m3api.PathNodeMsg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.m3api.PathNodeMsg}
 */
proto.m3api.PathNodeMsg.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setPathNodeId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setPathCtxId(value);
      break;
    case 3:
      var value = new m3point_pb.PointMsg;
      reader.readMessage(value,m3point_pb.PointMsg.deserializeBinaryFromReader);
      msg.setPoint(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setD(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setTrioId(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setConnectionMask(value);
      break;
    case 7:
      var value = /** @type {!Array<number>} */ (reader.readPackedInt64());
      msg.setLinkedPathNodeIdsList(value);
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
proto.m3api.PathNodeMsg.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.m3api.PathNodeMsg.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.m3api.PathNodeMsg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.PathNodeMsg.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPathNodeId();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getPathCtxId();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getPoint();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      m3point_pb.PointMsg.serializeBinaryToWriter
    );
  }
  f = message.getD();
  if (f !== 0) {
    writer.writeInt64(
      4,
      f
    );
  }
  f = message.getTrioId();
  if (f !== 0) {
    writer.writeInt32(
      5,
      f
    );
  }
  f = message.getConnectionMask();
  if (f !== 0) {
    writer.writeUint32(
      6,
      f
    );
  }
  f = message.getLinkedPathNodeIdsList();
  if (f.length > 0) {
    writer.writePackedInt64(
      7,
      f
    );
  }
};


/**
 * optional int64 path_node_id = 1;
 * @return {number}
 */
proto.m3api.PathNodeMsg.prototype.getPathNodeId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setPathNodeId = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int32 path_ctx_id = 2;
 * @return {number}
 */
proto.m3api.PathNodeMsg.prototype.getPathCtxId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setPathCtxId = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional PointMsg point = 3;
 * @return {?proto.m3api.PointMsg}
 */
proto.m3api.PathNodeMsg.prototype.getPoint = function() {
  return /** @type{?proto.m3api.PointMsg} */ (
    jspb.Message.getWrapperField(this, m3point_pb.PointMsg, 3));
};


/**
 * @param {?proto.m3api.PointMsg|undefined} value
 * @return {!proto.m3api.PathNodeMsg} returns this
*/
proto.m3api.PathNodeMsg.prototype.setPoint = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.clearPoint = function() {
  return this.setPoint(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.m3api.PathNodeMsg.prototype.hasPoint = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional int64 d = 4;
 * @return {number}
 */
proto.m3api.PathNodeMsg.prototype.getD = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setD = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional int32 trio_id = 5;
 * @return {number}
 */
proto.m3api.PathNodeMsg.prototype.getTrioId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setTrioId = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional uint32 connection_mask = 6;
 * @return {number}
 */
proto.m3api.PathNodeMsg.prototype.getConnectionMask = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setConnectionMask = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * repeated int64 linked_path_node_ids = 7;
 * @return {!Array<number>}
 */
proto.m3api.PathNodeMsg.prototype.getLinkedPathNodeIdsList = function() {
  return /** @type {!Array<number>} */ (jspb.Message.getRepeatedField(this, 7));
};


/**
 * @param {!Array<number>} value
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.setLinkedPathNodeIdsList = function(value) {
  return jspb.Message.setField(this, 7, value || []);
};


/**
 * @param {number} value
 * @param {number=} opt_index
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.addLinkedPathNodeIds = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 7, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.m3api.PathNodeMsg} returns this
 */
proto.m3api.PathNodeMsg.prototype.clearLinkedPathNodeIdsList = function() {
  return this.setLinkedPathNodeIdsList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.m3api.NextMoveRespMsg.repeatedFields_ = [1];



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
proto.m3api.NextMoveRespMsg.prototype.toObject = function(opt_includeInstance) {
  return proto.m3api.NextMoveRespMsg.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.m3api.NextMoveRespMsg} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.NextMoveRespMsg.toObject = function(includeInstance, msg) {
  var f, obj = {
    pathNodesList: jspb.Message.toObjectList(msg.getPathNodesList(),
    proto.m3api.PathNodeMsg.toObject, includeInstance)
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
 * @return {!proto.m3api.NextMoveRespMsg}
 */
proto.m3api.NextMoveRespMsg.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.m3api.NextMoveRespMsg;
  return proto.m3api.NextMoveRespMsg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.m3api.NextMoveRespMsg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.m3api.NextMoveRespMsg}
 */
proto.m3api.NextMoveRespMsg.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.m3api.PathNodeMsg;
      reader.readMessage(value,proto.m3api.PathNodeMsg.deserializeBinaryFromReader);
      msg.addPathNodes(value);
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
proto.m3api.NextMoveRespMsg.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.m3api.NextMoveRespMsg.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.m3api.NextMoveRespMsg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.m3api.NextMoveRespMsg.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPathNodesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.m3api.PathNodeMsg.serializeBinaryToWriter
    );
  }
};


/**
 * repeated PathNodeMsg path_nodes = 1;
 * @return {!Array<!proto.m3api.PathNodeMsg>}
 */
proto.m3api.NextMoveRespMsg.prototype.getPathNodesList = function() {
  return /** @type{!Array<!proto.m3api.PathNodeMsg>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.m3api.PathNodeMsg, 1));
};


/**
 * @param {!Array<!proto.m3api.PathNodeMsg>} value
 * @return {!proto.m3api.NextMoveRespMsg} returns this
*/
proto.m3api.NextMoveRespMsg.prototype.setPathNodesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.m3api.PathNodeMsg=} opt_value
 * @param {number=} opt_index
 * @return {!proto.m3api.PathNodeMsg}
 */
proto.m3api.NextMoveRespMsg.prototype.addPathNodes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.m3api.PathNodeMsg, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.m3api.NextMoveRespMsg} returns this
 */
proto.m3api.NextMoveRespMsg.prototype.clearPathNodesList = function() {
  return this.setPathNodesList([]);
};


goog.object.extend(exports, proto.m3api);
