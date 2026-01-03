// coverage:ignore-file
// ignore_for_file: type=lint

import 'package:json_annotation/json_annotation.dart';
import 'package:collection/collection.dart';

enum FieldType {
  @JsonValue(null)
  swaggerGeneratedUnknown(null),

  @JsonValue('string')
  string('string'),
  @JsonValue('integer')
  integer('integer'),
  @JsonValue('float')
  float('float'),
  @JsonValue('boolean')
  boolean('boolean'),
  @JsonValue('select')
  select('select'),
  @JsonValue('multiselect')
  multiselect('multiselect'),
  @JsonValue('ip_address')
  ipAddress('ip_address'),
  @JsonValue('ip_list')
  ipList('ip_list'),
  @JsonValue('regex')
  regex('regex'),
  @JsonValue('port_range')
  portRange('port_range');

  final String? value;

  const FieldType(this.value);
}

enum InputOutputType {
  @JsonValue(null)
  swaggerGeneratedUnknown(null),

  @JsonValue('http')
  http('http'),
  @JsonValue('ip')
  ip('ip');

  final String? value;

  const InputOutputType(this.value);
}

enum NodeStatus {
  @JsonValue(null)
  swaggerGeneratedUnknown(null),

  @JsonValue('Pending')
  pending('Pending'),
  @JsonValue('Starting')
  starting('Starting'),
  @JsonValue('Running')
  running('Running'),
  @JsonValue('Paused')
  paused('Paused'),
  @JsonValue('Stopped')
  stopped('Stopped'),
  @JsonValue('Completed')
  completed('Completed'),
  @JsonValue('Restarting')
  restarting('Restarting'),
  @JsonValue('OOMKilled')
  oomkilled('OOMKilled'),
  @JsonValue('Dead')
  dead('Dead'),
  @JsonValue('Error')
  error('Error'),
  @JsonValue('Unknown')
  unknown('Unknown');

  final String? value;

  const NodeStatus(this.value);
}

enum NodeType {
  @JsonValue(null)
  swaggerGeneratedUnknown(null),

  @JsonValue('trigger')
  trigger('trigger'),
  @JsonValue('worker')
  worker('worker');

  final String? value;

  const NodeType(this.value);
}

enum WorkerCategory {
  @JsonValue(null)
  swaggerGeneratedUnknown(null),

  @JsonValue('scanner')
  scanner('scanner'),
  @JsonValue('analyzer')
  analyzer('analyzer'),
  @JsonValue('utility')
  utility('utility'),
  @JsonValue('output')
  output('output');

  final String? value;

  const WorkerCategory(this.value);
}
