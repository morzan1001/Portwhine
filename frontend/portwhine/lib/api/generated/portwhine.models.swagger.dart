// coverage:ignore-file
// ignore_for_file: type=lint

import 'package:json_annotation/json_annotation.dart';
import 'package:collection/collection.dart';
import 'dart:convert';

import 'portwhine.enums.swagger.dart' as enums;

part 'portwhine.models.swagger.g.dart';

@JsonSerializable(explicitToJson: true)
class DeleteResponse {
  const DeleteResponse({this.detail});

  factory DeleteResponse.fromJson(Map<String, dynamic> json) =>
      _$DeleteResponseFromJson(json);

  static const toJsonFactory = _$DeleteResponseToJson;
  Map<String, dynamic> toJson() => _$DeleteResponseToJson(this);

  @JsonKey(name: 'detail')
  final String? detail;
  static const fromJsonFactory = _$DeleteResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is DeleteResponse &&
            (identical(other.detail, detail) ||
                const DeepCollectionEquality().equals(other.detail, detail)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(detail) ^ runtimeType.hashCode;
}

extension $DeleteResponseExtension on DeleteResponse {
  DeleteResponse copyWith({String? detail}) {
    return DeleteResponse(detail: detail ?? this.detail);
  }

  DeleteResponse copyWithWrapped({Wrapped<String?>? detail}) {
    return DeleteResponse(
      detail: (detail != null ? detail.value : this.detail),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class Edge {
  const Edge({
    required this.source,
    required this.target,
    this.sourcePort,
    this.targetPort,
  });

  factory Edge.fromJson(Map<String, dynamic> json) => _$EdgeFromJson(json);

  static const toJsonFactory = _$EdgeToJson;
  Map<String, dynamic> toJson() => _$EdgeToJson(this);

  @JsonKey(name: 'source')
  final String source;
  @JsonKey(name: 'target')
  final String target;
  @JsonKey(name: 'source_port')
  final String? sourcePort;
  @JsonKey(name: 'target_port')
  final String? targetPort;
  static const fromJsonFactory = _$EdgeFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is Edge &&
            (identical(other.source, source) ||
                const DeepCollectionEquality().equals(other.source, source)) &&
            (identical(other.target, target) ||
                const DeepCollectionEquality().equals(other.target, target)) &&
            (identical(other.sourcePort, sourcePort) ||
                const DeepCollectionEquality().equals(
                  other.sourcePort,
                  sourcePort,
                )) &&
            (identical(other.targetPort, targetPort) ||
                const DeepCollectionEquality().equals(
                  other.targetPort,
                  targetPort,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(source) ^
      const DeepCollectionEquality().hash(target) ^
      const DeepCollectionEquality().hash(sourcePort) ^
      const DeepCollectionEquality().hash(targetPort) ^
      runtimeType.hashCode;
}

extension $EdgeExtension on Edge {
  Edge copyWith({
    String? source,
    String? target,
    String? sourcePort,
    String? targetPort,
  }) {
    return Edge(
      source: source ?? this.source,
      target: target ?? this.target,
      sourcePort: sourcePort ?? this.sourcePort,
      targetPort: targetPort ?? this.targetPort,
    );
  }

  Edge copyWithWrapped({
    Wrapped<String>? source,
    Wrapped<String>? target,
    Wrapped<String?>? sourcePort,
    Wrapped<String?>? targetPort,
  }) {
    return Edge(
      source: (source != null ? source.value : this.source),
      target: (target != null ? target.value : this.target),
      sourcePort: (sourcePort != null ? sourcePort.value : this.sourcePort),
      targetPort: (targetPort != null ? targetPort.value : this.targetPort),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class FieldDefinition {
  const FieldDefinition({
    required this.name,
    required this.label,
    required this.type,
    this.description,
    this.required,
    this.$default,
    this.options,
    this.placeholder,
    this.validationPattern,
    this.minValue,
    this.maxValue,
  });

  factory FieldDefinition.fromJson(Map<String, dynamic> json) =>
      _$FieldDefinitionFromJson(json);

  static const toJsonFactory = _$FieldDefinitionToJson;
  Map<String, dynamic> toJson() => _$FieldDefinitionToJson(this);

  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'label')
  final String label;
  @JsonKey(name: 'type', toJson: fieldTypeToJson, fromJson: fieldTypeFromJson)
  final enums.FieldType type;
  @JsonKey(name: 'description')
  final String? description;
  @JsonKey(name: 'required', defaultValue: false)
  final bool? required;
  @JsonKey(name: 'default')
  final Object? $default;
  @JsonKey(name: 'options', defaultValue: <String>[])
  final List<String>? options;
  @JsonKey(name: 'placeholder')
  final String? placeholder;
  @JsonKey(name: 'validation_pattern')
  final String? validationPattern;
  @JsonKey(name: 'min_value')
  final double? minValue;
  @JsonKey(name: 'max_value')
  final double? maxValue;
  static const fromJsonFactory = _$FieldDefinitionFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is FieldDefinition &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.label, label) ||
                const DeepCollectionEquality().equals(other.label, label)) &&
            (identical(other.type, type) ||
                const DeepCollectionEquality().equals(other.type, type)) &&
            (identical(other.description, description) ||
                const DeepCollectionEquality().equals(
                  other.description,
                  description,
                )) &&
            (identical(other.required, required) ||
                const DeepCollectionEquality().equals(
                  other.required,
                  required,
                )) &&
            (identical(other.$default, $default) ||
                const DeepCollectionEquality().equals(
                  other.$default,
                  $default,
                )) &&
            (identical(other.options, options) ||
                const DeepCollectionEquality().equals(
                  other.options,
                  options,
                )) &&
            (identical(other.placeholder, placeholder) ||
                const DeepCollectionEquality().equals(
                  other.placeholder,
                  placeholder,
                )) &&
            (identical(other.validationPattern, validationPattern) ||
                const DeepCollectionEquality().equals(
                  other.validationPattern,
                  validationPattern,
                )) &&
            (identical(other.minValue, minValue) ||
                const DeepCollectionEquality().equals(
                  other.minValue,
                  minValue,
                )) &&
            (identical(other.maxValue, maxValue) ||
                const DeepCollectionEquality().equals(
                  other.maxValue,
                  maxValue,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(label) ^
      const DeepCollectionEquality().hash(type) ^
      const DeepCollectionEquality().hash(description) ^
      const DeepCollectionEquality().hash(required) ^
      const DeepCollectionEquality().hash($default) ^
      const DeepCollectionEquality().hash(options) ^
      const DeepCollectionEquality().hash(placeholder) ^
      const DeepCollectionEquality().hash(validationPattern) ^
      const DeepCollectionEquality().hash(minValue) ^
      const DeepCollectionEquality().hash(maxValue) ^
      runtimeType.hashCode;
}

extension $FieldDefinitionExtension on FieldDefinition {
  FieldDefinition copyWith({
    String? name,
    String? label,
    enums.FieldType? type,
    String? description,
    bool? required,
    Object? $default,
    List<String>? options,
    String? placeholder,
    String? validationPattern,
    double? minValue,
    double? maxValue,
  }) {
    return FieldDefinition(
      name: name ?? this.name,
      label: label ?? this.label,
      type: type ?? this.type,
      description: description ?? this.description,
      required: required ?? this.required,
      $default: $default ?? this.$default,
      options: options ?? this.options,
      placeholder: placeholder ?? this.placeholder,
      validationPattern: validationPattern ?? this.validationPattern,
      minValue: minValue ?? this.minValue,
      maxValue: maxValue ?? this.maxValue,
    );
  }

  FieldDefinition copyWithWrapped({
    Wrapped<String>? name,
    Wrapped<String>? label,
    Wrapped<enums.FieldType>? type,
    Wrapped<String?>? description,
    Wrapped<bool?>? required,
    Wrapped<Object?>? $default,
    Wrapped<List<String>?>? options,
    Wrapped<String?>? placeholder,
    Wrapped<String?>? validationPattern,
    Wrapped<double?>? minValue,
    Wrapped<double?>? maxValue,
  }) {
    return FieldDefinition(
      name: (name != null ? name.value : this.name),
      label: (label != null ? label.value : this.label),
      type: (type != null ? type.value : this.type),
      description: (description != null ? description.value : this.description),
      required: (required != null ? required.value : this.required),
      $default: ($default != null ? $default.value : this.$default),
      options: (options != null ? options.value : this.options),
      placeholder: (placeholder != null ? placeholder.value : this.placeholder),
      validationPattern: (validationPattern != null
          ? validationPattern.value
          : this.validationPattern),
      minValue: (minValue != null ? minValue.value : this.minValue),
      maxValue: (maxValue != null ? maxValue.value : this.maxValue),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class GridPosition {
  const GridPosition({this.x, this.y});

  factory GridPosition.fromJson(Map<String, dynamic> json) =>
      _$GridPositionFromJson(json);

  static const toJsonFactory = _$GridPositionToJson;
  Map<String, dynamic> toJson() => _$GridPositionToJson(this);

  @JsonKey(name: 'x')
  final double? x;
  @JsonKey(name: 'y')
  final double? y;
  static const fromJsonFactory = _$GridPositionFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is GridPosition &&
            (identical(other.x, x) ||
                const DeepCollectionEquality().equals(other.x, x)) &&
            (identical(other.y, y) ||
                const DeepCollectionEquality().equals(other.y, y)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(x) ^
      const DeepCollectionEquality().hash(y) ^
      runtimeType.hashCode;
}

extension $GridPositionExtension on GridPosition {
  GridPosition copyWith({double? x, double? y}) {
    return GridPosition(x: x ?? this.x, y: y ?? this.y);
  }

  GridPosition copyWithWrapped({Wrapped<double?>? x, Wrapped<double?>? y}) {
    return GridPosition(
      x: (x != null ? x.value : this.x),
      y: (y != null ? y.value : this.y),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class HTTPValidationError {
  const HTTPValidationError({this.detail});

  factory HTTPValidationError.fromJson(Map<String, dynamic> json) =>
      _$HTTPValidationErrorFromJson(json);

  static const toJsonFactory = _$HTTPValidationErrorToJson;
  Map<String, dynamic> toJson() => _$HTTPValidationErrorToJson(this);

  @JsonKey(name: 'detail', defaultValue: <ValidationError>[])
  final List<ValidationError>? detail;
  static const fromJsonFactory = _$HTTPValidationErrorFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is HTTPValidationError &&
            (identical(other.detail, detail) ||
                const DeepCollectionEquality().equals(other.detail, detail)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(detail) ^ runtimeType.hashCode;
}

extension $HTTPValidationErrorExtension on HTTPValidationError {
  HTTPValidationError copyWith({List<ValidationError>? detail}) {
    return HTTPValidationError(detail: detail ?? this.detail);
  }

  HTTPValidationError copyWithWrapped({
    Wrapped<List<ValidationError>?>? detail,
  }) {
    return HTTPValidationError(
      detail: (detail != null ? detail.value : this.detail),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class HttpTarget {
  const HttpTarget({required this.url, this.method, this.headers});

  factory HttpTarget.fromJson(Map<String, dynamic> json) =>
      _$HttpTargetFromJson(json);

  static const toJsonFactory = _$HttpTargetToJson;
  Map<String, dynamic> toJson() => _$HttpTargetToJson(this);

  @JsonKey(name: 'url')
  final String url;
  @JsonKey(name: 'method')
  final String? method;
  @JsonKey(name: 'headers')
  final Map<String, dynamic>? headers;
  static const fromJsonFactory = _$HttpTargetFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is HttpTarget &&
            (identical(other.url, url) ||
                const DeepCollectionEquality().equals(other.url, url)) &&
            (identical(other.method, method) ||
                const DeepCollectionEquality().equals(other.method, method)) &&
            (identical(other.headers, headers) ||
                const DeepCollectionEquality().equals(other.headers, headers)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(url) ^
      const DeepCollectionEquality().hash(method) ^
      const DeepCollectionEquality().hash(headers) ^
      runtimeType.hashCode;
}

extension $HttpTargetExtension on HttpTarget {
  HttpTarget copyWith({
    String? url,
    String? method,
    Map<String, dynamic>? headers,
  }) {
    return HttpTarget(
      url: url ?? this.url,
      method: method ?? this.method,
      headers: headers ?? this.headers,
    );
  }

  HttpTarget copyWithWrapped({
    Wrapped<String>? url,
    Wrapped<String?>? method,
    Wrapped<Map<String, dynamic>?>? headers,
  }) {
    return HttpTarget(
      url: (url != null ? url.value : this.url),
      method: (method != null ? method.value : this.method),
      headers: (headers != null ? headers.value : this.headers),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class InstanceHealth {
  const InstanceHealth({
    required this.instanceNumber,
    this.containerId,
    this.containerName,
    this.status,
    this.startedAt,
    this.finishedAt,
    this.exitCode,
    this.errorMessage,
    this.jobsProcessed,
  });

  factory InstanceHealth.fromJson(Map<String, dynamic> json) =>
      _$InstanceHealthFromJson(json);

  static const toJsonFactory = _$InstanceHealthToJson;
  Map<String, dynamic> toJson() => _$InstanceHealthToJson(this);

  @JsonKey(name: 'instance_number')
  final int instanceNumber;
  @JsonKey(name: 'container_id')
  final String? containerId;
  @JsonKey(name: 'container_name')
  final String? containerName;
  @JsonKey(
    name: 'status',
    toJson: nodeStatusNullableToJson,
    fromJson: nodeStatusStatusNullableFromJson,
  )
  final enums.NodeStatus? status;
  static enums.NodeStatus? nodeStatusStatusNullableFromJson(Object? value) =>
      nodeStatusNullableFromJson(value, enums.NodeStatus.unknown);

  @JsonKey(name: 'started_at')
  final DateTime? startedAt;
  @JsonKey(name: 'finished_at')
  final DateTime? finishedAt;
  @JsonKey(name: 'exit_code')
  final int? exitCode;
  @JsonKey(name: 'error_message')
  final String? errorMessage;
  @JsonKey(name: 'jobs_processed')
  final int? jobsProcessed;
  static const fromJsonFactory = _$InstanceHealthFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is InstanceHealth &&
            (identical(other.instanceNumber, instanceNumber) ||
                const DeepCollectionEquality().equals(
                  other.instanceNumber,
                  instanceNumber,
                )) &&
            (identical(other.containerId, containerId) ||
                const DeepCollectionEquality().equals(
                  other.containerId,
                  containerId,
                )) &&
            (identical(other.containerName, containerName) ||
                const DeepCollectionEquality().equals(
                  other.containerName,
                  containerName,
                )) &&
            (identical(other.status, status) ||
                const DeepCollectionEquality().equals(other.status, status)) &&
            (identical(other.startedAt, startedAt) ||
                const DeepCollectionEquality().equals(
                  other.startedAt,
                  startedAt,
                )) &&
            (identical(other.finishedAt, finishedAt) ||
                const DeepCollectionEquality().equals(
                  other.finishedAt,
                  finishedAt,
                )) &&
            (identical(other.exitCode, exitCode) ||
                const DeepCollectionEquality().equals(
                  other.exitCode,
                  exitCode,
                )) &&
            (identical(other.errorMessage, errorMessage) ||
                const DeepCollectionEquality().equals(
                  other.errorMessage,
                  errorMessage,
                )) &&
            (identical(other.jobsProcessed, jobsProcessed) ||
                const DeepCollectionEquality().equals(
                  other.jobsProcessed,
                  jobsProcessed,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(instanceNumber) ^
      const DeepCollectionEquality().hash(containerId) ^
      const DeepCollectionEquality().hash(containerName) ^
      const DeepCollectionEquality().hash(status) ^
      const DeepCollectionEquality().hash(startedAt) ^
      const DeepCollectionEquality().hash(finishedAt) ^
      const DeepCollectionEquality().hash(exitCode) ^
      const DeepCollectionEquality().hash(errorMessage) ^
      const DeepCollectionEquality().hash(jobsProcessed) ^
      runtimeType.hashCode;
}

extension $InstanceHealthExtension on InstanceHealth {
  InstanceHealth copyWith({
    int? instanceNumber,
    String? containerId,
    String? containerName,
    enums.NodeStatus? status,
    DateTime? startedAt,
    DateTime? finishedAt,
    int? exitCode,
    String? errorMessage,
    int? jobsProcessed,
  }) {
    return InstanceHealth(
      instanceNumber: instanceNumber ?? this.instanceNumber,
      containerId: containerId ?? this.containerId,
      containerName: containerName ?? this.containerName,
      status: status ?? this.status,
      startedAt: startedAt ?? this.startedAt,
      finishedAt: finishedAt ?? this.finishedAt,
      exitCode: exitCode ?? this.exitCode,
      errorMessage: errorMessage ?? this.errorMessage,
      jobsProcessed: jobsProcessed ?? this.jobsProcessed,
    );
  }

  InstanceHealth copyWithWrapped({
    Wrapped<int>? instanceNumber,
    Wrapped<String?>? containerId,
    Wrapped<String?>? containerName,
    Wrapped<enums.NodeStatus?>? status,
    Wrapped<DateTime?>? startedAt,
    Wrapped<DateTime?>? finishedAt,
    Wrapped<int?>? exitCode,
    Wrapped<String?>? errorMessage,
    Wrapped<int?>? jobsProcessed,
  }) {
    return InstanceHealth(
      instanceNumber: (instanceNumber != null
          ? instanceNumber.value
          : this.instanceNumber),
      containerId: (containerId != null ? containerId.value : this.containerId),
      containerName: (containerName != null
          ? containerName.value
          : this.containerName),
      status: (status != null ? status.value : this.status),
      startedAt: (startedAt != null ? startedAt.value : this.startedAt),
      finishedAt: (finishedAt != null ? finishedAt.value : this.finishedAt),
      exitCode: (exitCode != null ? exitCode.value : this.exitCode),
      errorMessage: (errorMessage != null
          ? errorMessage.value
          : this.errorMessage),
      jobsProcessed: (jobsProcessed != null
          ? jobsProcessed.value
          : this.jobsProcessed),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class InstanceHealthResponse {
  const InstanceHealthResponse({
    required this.instanceNumber,
    this.containerId,
    this.containerName,
    this.status,
    this.startedAt,
    this.finishedAt,
    this.exitCode,
    this.errorMessage,
    this.jobsProcessed,
  });

  factory InstanceHealthResponse.fromJson(Map<String, dynamic> json) =>
      _$InstanceHealthResponseFromJson(json);

  static const toJsonFactory = _$InstanceHealthResponseToJson;
  Map<String, dynamic> toJson() => _$InstanceHealthResponseToJson(this);

  @JsonKey(name: 'instance_number')
  final int instanceNumber;
  @JsonKey(name: 'container_id')
  final String? containerId;
  @JsonKey(name: 'container_name')
  final String? containerName;
  @JsonKey(name: 'status')
  final String? status;
  @JsonKey(name: 'started_at')
  final String? startedAt;
  @JsonKey(name: 'finished_at')
  final String? finishedAt;
  @JsonKey(name: 'exit_code')
  final int? exitCode;
  @JsonKey(name: 'error_message')
  final String? errorMessage;
  @JsonKey(name: 'jobs_processed')
  final int? jobsProcessed;
  static const fromJsonFactory = _$InstanceHealthResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is InstanceHealthResponse &&
            (identical(other.instanceNumber, instanceNumber) ||
                const DeepCollectionEquality().equals(
                  other.instanceNumber,
                  instanceNumber,
                )) &&
            (identical(other.containerId, containerId) ||
                const DeepCollectionEquality().equals(
                  other.containerId,
                  containerId,
                )) &&
            (identical(other.containerName, containerName) ||
                const DeepCollectionEquality().equals(
                  other.containerName,
                  containerName,
                )) &&
            (identical(other.status, status) ||
                const DeepCollectionEquality().equals(other.status, status)) &&
            (identical(other.startedAt, startedAt) ||
                const DeepCollectionEquality().equals(
                  other.startedAt,
                  startedAt,
                )) &&
            (identical(other.finishedAt, finishedAt) ||
                const DeepCollectionEquality().equals(
                  other.finishedAt,
                  finishedAt,
                )) &&
            (identical(other.exitCode, exitCode) ||
                const DeepCollectionEquality().equals(
                  other.exitCode,
                  exitCode,
                )) &&
            (identical(other.errorMessage, errorMessage) ||
                const DeepCollectionEquality().equals(
                  other.errorMessage,
                  errorMessage,
                )) &&
            (identical(other.jobsProcessed, jobsProcessed) ||
                const DeepCollectionEquality().equals(
                  other.jobsProcessed,
                  jobsProcessed,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(instanceNumber) ^
      const DeepCollectionEquality().hash(containerId) ^
      const DeepCollectionEquality().hash(containerName) ^
      const DeepCollectionEquality().hash(status) ^
      const DeepCollectionEquality().hash(startedAt) ^
      const DeepCollectionEquality().hash(finishedAt) ^
      const DeepCollectionEquality().hash(exitCode) ^
      const DeepCollectionEquality().hash(errorMessage) ^
      const DeepCollectionEquality().hash(jobsProcessed) ^
      runtimeType.hashCode;
}

extension $InstanceHealthResponseExtension on InstanceHealthResponse {
  InstanceHealthResponse copyWith({
    int? instanceNumber,
    String? containerId,
    String? containerName,
    String? status,
    String? startedAt,
    String? finishedAt,
    int? exitCode,
    String? errorMessage,
    int? jobsProcessed,
  }) {
    return InstanceHealthResponse(
      instanceNumber: instanceNumber ?? this.instanceNumber,
      containerId: containerId ?? this.containerId,
      containerName: containerName ?? this.containerName,
      status: status ?? this.status,
      startedAt: startedAt ?? this.startedAt,
      finishedAt: finishedAt ?? this.finishedAt,
      exitCode: exitCode ?? this.exitCode,
      errorMessage: errorMessage ?? this.errorMessage,
      jobsProcessed: jobsProcessed ?? this.jobsProcessed,
    );
  }

  InstanceHealthResponse copyWithWrapped({
    Wrapped<int>? instanceNumber,
    Wrapped<String?>? containerId,
    Wrapped<String?>? containerName,
    Wrapped<String?>? status,
    Wrapped<String?>? startedAt,
    Wrapped<String?>? finishedAt,
    Wrapped<int?>? exitCode,
    Wrapped<String?>? errorMessage,
    Wrapped<int?>? jobsProcessed,
  }) {
    return InstanceHealthResponse(
      instanceNumber: (instanceNumber != null
          ? instanceNumber.value
          : this.instanceNumber),
      containerId: (containerId != null ? containerId.value : this.containerId),
      containerName: (containerName != null
          ? containerName.value
          : this.containerName),
      status: (status != null ? status.value : this.status),
      startedAt: (startedAt != null ? startedAt.value : this.startedAt),
      finishedAt: (finishedAt != null ? finishedAt.value : this.finishedAt),
      exitCode: (exitCode != null ? exitCode.value : this.exitCode),
      errorMessage: (errorMessage != null
          ? errorMessage.value
          : this.errorMessage),
      jobsProcessed: (jobsProcessed != null
          ? jobsProcessed.value
          : this.jobsProcessed),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class IpTarget {
  const IpTarget({required this.ip, this.port});

  factory IpTarget.fromJson(Map<String, dynamic> json) =>
      _$IpTargetFromJson(json);

  static const toJsonFactory = _$IpTargetToJson;
  Map<String, dynamic> toJson() => _$IpTargetToJson(this);

  @JsonKey(name: 'ip')
  final dynamic ip;
  @JsonKey(name: 'port')
  final int? port;
  static const fromJsonFactory = _$IpTargetFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is IpTarget &&
            (identical(other.ip, ip) ||
                const DeepCollectionEquality().equals(other.ip, ip)) &&
            (identical(other.port, port) ||
                const DeepCollectionEquality().equals(other.port, port)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(ip) ^
      const DeepCollectionEquality().hash(port) ^
      runtimeType.hashCode;
}

extension $IpTargetExtension on IpTarget {
  IpTarget copyWith({dynamic ip, int? port}) {
    return IpTarget(ip: ip ?? this.ip, port: port ?? this.port);
  }

  IpTarget copyWithWrapped({Wrapped<dynamic>? ip, Wrapped<int?>? port}) {
    return IpTarget(
      ip: (ip != null ? ip.value : this.ip),
      port: (port != null ? port.value : this.port),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class JobPayload {
  const JobPayload({this.http, this.ip});

  factory JobPayload.fromJson(Map<String, dynamic> json) =>
      _$JobPayloadFromJson(json);

  static const toJsonFactory = _$JobPayloadToJson;
  Map<String, dynamic> toJson() => _$JobPayloadToJson(this);

  @JsonKey(name: 'http', defaultValue: <HttpTarget>[])
  final List<HttpTarget>? http;
  @JsonKey(name: 'ip', defaultValue: <IpTarget>[])
  final List<IpTarget>? ip;
  static const fromJsonFactory = _$JobPayloadFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is JobPayload &&
            (identical(other.http, http) ||
                const DeepCollectionEquality().equals(other.http, http)) &&
            (identical(other.ip, ip) ||
                const DeepCollectionEquality().equals(other.ip, ip)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(http) ^
      const DeepCollectionEquality().hash(ip) ^
      runtimeType.hashCode;
}

extension $JobPayloadExtension on JobPayload {
  JobPayload copyWith({List<HttpTarget>? http, List<IpTarget>? ip}) {
    return JobPayload(http: http ?? this.http, ip: ip ?? this.ip);
  }

  JobPayload copyWithWrapped({
    Wrapped<List<HttpTarget>?>? http,
    Wrapped<List<IpTarget>?>? ip,
  }) {
    return JobPayload(
      http: (http != null ? http.value : this.http),
      ip: (ip != null ? ip.value : this.ip),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class MessageResponse {
  const MessageResponse({required this.detail});

  factory MessageResponse.fromJson(Map<String, dynamic> json) =>
      _$MessageResponseFromJson(json);

  static const toJsonFactory = _$MessageResponseToJson;
  Map<String, dynamic> toJson() => _$MessageResponseToJson(this);

  @JsonKey(name: 'detail')
  final String detail;
  static const fromJsonFactory = _$MessageResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is MessageResponse &&
            (identical(other.detail, detail) ||
                const DeepCollectionEquality().equals(other.detail, detail)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(detail) ^ runtimeType.hashCode;
}

extension $MessageResponseExtension on MessageResponse {
  MessageResponse copyWith({String? detail}) {
    return MessageResponse(detail: detail ?? this.detail);
  }

  MessageResponse copyWithWrapped({Wrapped<String>? detail}) {
    return MessageResponse(
      detail: (detail != null ? detail.value : this.detail),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class NodeConfigExampleResponse {
  const NodeConfigExampleResponse({
    required this.description,
    required this.example,
  });

  factory NodeConfigExampleResponse.fromJson(Map<String, dynamic> json) =>
      _$NodeConfigExampleResponseFromJson(json);

  static const toJsonFactory = _$NodeConfigExampleResponseToJson;
  Map<String, dynamic> toJson() => _$NodeConfigExampleResponseToJson(this);

  @JsonKey(name: 'description')
  final String description;
  @JsonKey(name: 'example')
  final Map<String, dynamic> example;
  static const fromJsonFactory = _$NodeConfigExampleResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is NodeConfigExampleResponse &&
            (identical(other.description, description) ||
                const DeepCollectionEquality().equals(
                  other.description,
                  description,
                )) &&
            (identical(other.example, example) ||
                const DeepCollectionEquality().equals(other.example, example)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(description) ^
      const DeepCollectionEquality().hash(example) ^
      runtimeType.hashCode;
}

extension $NodeConfigExampleResponseExtension on NodeConfigExampleResponse {
  NodeConfigExampleResponse copyWith({
    String? description,
    Map<String, dynamic>? example,
  }) {
    return NodeConfigExampleResponse(
      description: description ?? this.description,
      example: example ?? this.example,
    );
  }

  NodeConfigExampleResponse copyWithWrapped({
    Wrapped<String>? description,
    Wrapped<Map<String, dynamic>>? example,
  }) {
    return NodeConfigExampleResponse(
      description: (description != null ? description.value : this.description),
      example: (example != null ? example.value : this.example),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class NodeDefinition {
  const NodeDefinition({
    required this.id,
    required this.name,
    required this.description,
    required this.nodeType,
    this.category,
    this.icon,
    this.color,
    this.inputs,
    this.outputs,
    this.configFields,
    required this.imageName,
    this.supportsMultipleInstances,
    this.maxInstances,
  });

  factory NodeDefinition.fromJson(Map<String, dynamic> json) =>
      _$NodeDefinitionFromJson(json);

  static const toJsonFactory = _$NodeDefinitionToJson;
  Map<String, dynamic> toJson() => _$NodeDefinitionToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'description')
  final String description;
  @JsonKey(
    name: 'node_type',
    toJson: nodeTypeToJson,
    fromJson: nodeTypeFromJson,
  )
  final enums.NodeType nodeType;
  @JsonKey(
    name: 'category',
    toJson: workerCategoryNullableToJson,
    fromJson: workerCategoryNullableFromJson,
  )
  final enums.WorkerCategory? category;
  @JsonKey(name: 'icon')
  final String? icon;
  @JsonKey(name: 'color')
  final String? color;
  @JsonKey(name: 'inputs', defaultValue: <PortDefinition>[])
  final List<PortDefinition>? inputs;
  @JsonKey(name: 'outputs', defaultValue: <PortDefinition>[])
  final List<PortDefinition>? outputs;
  @JsonKey(name: 'config_fields', defaultValue: <FieldDefinition>[])
  final List<FieldDefinition>? configFields;
  @JsonKey(name: 'image_name')
  final String imageName;
  @JsonKey(name: 'supports_multiple_instances', defaultValue: true)
  final bool? supportsMultipleInstances;
  @JsonKey(name: 'max_instances')
  final int? maxInstances;
  static const fromJsonFactory = _$NodeDefinitionFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is NodeDefinition &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.description, description) ||
                const DeepCollectionEquality().equals(
                  other.description,
                  description,
                )) &&
            (identical(other.nodeType, nodeType) ||
                const DeepCollectionEquality().equals(
                  other.nodeType,
                  nodeType,
                )) &&
            (identical(other.category, category) ||
                const DeepCollectionEquality().equals(
                  other.category,
                  category,
                )) &&
            (identical(other.icon, icon) ||
                const DeepCollectionEquality().equals(other.icon, icon)) &&
            (identical(other.color, color) ||
                const DeepCollectionEquality().equals(other.color, color)) &&
            (identical(other.inputs, inputs) ||
                const DeepCollectionEquality().equals(other.inputs, inputs)) &&
            (identical(other.outputs, outputs) ||
                const DeepCollectionEquality().equals(
                  other.outputs,
                  outputs,
                )) &&
            (identical(other.configFields, configFields) ||
                const DeepCollectionEquality().equals(
                  other.configFields,
                  configFields,
                )) &&
            (identical(other.imageName, imageName) ||
                const DeepCollectionEquality().equals(
                  other.imageName,
                  imageName,
                )) &&
            (identical(
                  other.supportsMultipleInstances,
                  supportsMultipleInstances,
                ) ||
                const DeepCollectionEquality().equals(
                  other.supportsMultipleInstances,
                  supportsMultipleInstances,
                )) &&
            (identical(other.maxInstances, maxInstances) ||
                const DeepCollectionEquality().equals(
                  other.maxInstances,
                  maxInstances,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(description) ^
      const DeepCollectionEquality().hash(nodeType) ^
      const DeepCollectionEquality().hash(category) ^
      const DeepCollectionEquality().hash(icon) ^
      const DeepCollectionEquality().hash(color) ^
      const DeepCollectionEquality().hash(inputs) ^
      const DeepCollectionEquality().hash(outputs) ^
      const DeepCollectionEquality().hash(configFields) ^
      const DeepCollectionEquality().hash(imageName) ^
      const DeepCollectionEquality().hash(supportsMultipleInstances) ^
      const DeepCollectionEquality().hash(maxInstances) ^
      runtimeType.hashCode;
}

extension $NodeDefinitionExtension on NodeDefinition {
  NodeDefinition copyWith({
    String? id,
    String? name,
    String? description,
    enums.NodeType? nodeType,
    enums.WorkerCategory? category,
    String? icon,
    String? color,
    List<PortDefinition>? inputs,
    List<PortDefinition>? outputs,
    List<FieldDefinition>? configFields,
    String? imageName,
    bool? supportsMultipleInstances,
    int? maxInstances,
  }) {
    return NodeDefinition(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      nodeType: nodeType ?? this.nodeType,
      category: category ?? this.category,
      icon: icon ?? this.icon,
      color: color ?? this.color,
      inputs: inputs ?? this.inputs,
      outputs: outputs ?? this.outputs,
      configFields: configFields ?? this.configFields,
      imageName: imageName ?? this.imageName,
      supportsMultipleInstances:
          supportsMultipleInstances ?? this.supportsMultipleInstances,
      maxInstances: maxInstances ?? this.maxInstances,
    );
  }

  NodeDefinition copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<String>? name,
    Wrapped<String>? description,
    Wrapped<enums.NodeType>? nodeType,
    Wrapped<enums.WorkerCategory?>? category,
    Wrapped<String?>? icon,
    Wrapped<String?>? color,
    Wrapped<List<PortDefinition>?>? inputs,
    Wrapped<List<PortDefinition>?>? outputs,
    Wrapped<List<FieldDefinition>?>? configFields,
    Wrapped<String>? imageName,
    Wrapped<bool?>? supportsMultipleInstances,
    Wrapped<int?>? maxInstances,
  }) {
    return NodeDefinition(
      id: (id != null ? id.value : this.id),
      name: (name != null ? name.value : this.name),
      description: (description != null ? description.value : this.description),
      nodeType: (nodeType != null ? nodeType.value : this.nodeType),
      category: (category != null ? category.value : this.category),
      icon: (icon != null ? icon.value : this.icon),
      color: (color != null ? color.value : this.color),
      inputs: (inputs != null ? inputs.value : this.inputs),
      outputs: (outputs != null ? outputs.value : this.outputs),
      configFields: (configFields != null
          ? configFields.value
          : this.configFields),
      imageName: (imageName != null ? imageName.value : this.imageName),
      supportsMultipleInstances: (supportsMultipleInstances != null
          ? supportsMultipleInstances.value
          : this.supportsMultipleInstances),
      maxInstances: (maxInstances != null
          ? maxInstances.value
          : this.maxInstances),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class Pipeline {
  const Pipeline({required this.name, this.trigger, this.worker, this.edges});

  factory Pipeline.fromJson(Map<String, dynamic> json) =>
      _$PipelineFromJson(json);

  static const toJsonFactory = _$PipelineToJson;
  Map<String, dynamic> toJson() => _$PipelineToJson(this);

  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'trigger')
  final TriggerConfig? trigger;
  @JsonKey(name: 'worker')
  final List<WorkerConfig>? worker;
  @JsonKey(name: 'edges', defaultValue: <Edge>[])
  final List<Edge>? edges;
  static const fromJsonFactory = _$PipelineFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is Pipeline &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.trigger, trigger) ||
                const DeepCollectionEquality().equals(
                  other.trigger,
                  trigger,
                )) &&
            (identical(other.worker, worker) ||
                const DeepCollectionEquality().equals(other.worker, worker)) &&
            (identical(other.edges, edges) ||
                const DeepCollectionEquality().equals(other.edges, edges)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(trigger) ^
      const DeepCollectionEquality().hash(worker) ^
      const DeepCollectionEquality().hash(edges) ^
      runtimeType.hashCode;
}

extension $PipelineExtension on Pipeline {
  Pipeline copyWith({
    String? name,
    TriggerConfig? trigger,
    List<WorkerConfig>? worker,
    List<Edge>? edges,
  }) {
    return Pipeline(
      name: name ?? this.name,
      trigger: trigger ?? this.trigger,
      worker: worker ?? this.worker,
      edges: edges ?? this.edges,
    );
  }

  Pipeline copyWithWrapped({
    Wrapped<String>? name,
    Wrapped<TriggerConfig?>? trigger,
    Wrapped<List<WorkerConfig>?>? worker,
    Wrapped<List<Edge>?>? edges,
  }) {
    return Pipeline(
      name: (name != null ? name.value : this.name),
      trigger: (trigger != null ? trigger.value : this.trigger),
      worker: (worker != null ? worker.value : this.worker),
      edges: (edges != null ? edges.value : this.edges),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class PipelineListItem {
  const PipelineListItem({required this.id, required this.name, this.status});

  factory PipelineListItem.fromJson(Map<String, dynamic> json) =>
      _$PipelineListItemFromJson(json);

  static const toJsonFactory = _$PipelineListItemToJson;
  Map<String, dynamic> toJson() => _$PipelineListItemToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'status')
  final String? status;
  static const fromJsonFactory = _$PipelineListItemFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PipelineListItem &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.status, status) ||
                const DeepCollectionEquality().equals(other.status, status)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(status) ^
      runtimeType.hashCode;
}

extension $PipelineListItemExtension on PipelineListItem {
  PipelineListItem copyWith({String? id, String? name, String? status}) {
    return PipelineListItem(
      id: id ?? this.id,
      name: name ?? this.name,
      status: status ?? this.status,
    );
  }

  PipelineListItem copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<String>? name,
    Wrapped<String?>? status,
  }) {
    return PipelineListItem(
      id: (id != null ? id.value : this.id),
      name: (name != null ? name.value : this.name),
      status: (status != null ? status.value : this.status),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class PipelinePatch {
  const PipelinePatch({
    required this.id,
    this.name,
    this.trigger,
    this.worker,
    this.edges,
  });

  factory PipelinePatch.fromJson(Map<String, dynamic> json) =>
      _$PipelinePatchFromJson(json);

  static const toJsonFactory = _$PipelinePatchToJson;
  Map<String, dynamic> toJson() => _$PipelinePatchToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'name')
  final String? name;
  @JsonKey(name: 'trigger')
  final Object? trigger;
  @JsonKey(name: 'worker')
  final List<Object>? worker;
  @JsonKey(name: 'edges')
  final List<Object>? edges;
  static const fromJsonFactory = _$PipelinePatchFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PipelinePatch &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.trigger, trigger) ||
                const DeepCollectionEquality().equals(
                  other.trigger,
                  trigger,
                )) &&
            (identical(other.worker, worker) ||
                const DeepCollectionEquality().equals(other.worker, worker)) &&
            (identical(other.edges, edges) ||
                const DeepCollectionEquality().equals(other.edges, edges)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(trigger) ^
      const DeepCollectionEquality().hash(worker) ^
      const DeepCollectionEquality().hash(edges) ^
      runtimeType.hashCode;
}

extension $PipelinePatchExtension on PipelinePatch {
  PipelinePatch copyWith({
    String? id,
    String? name,
    Object? trigger,
    List<Object>? worker,
    List<Object>? edges,
  }) {
    return PipelinePatch(
      id: id ?? this.id,
      name: name ?? this.name,
      trigger: trigger ?? this.trigger,
      worker: worker ?? this.worker,
      edges: edges ?? this.edges,
    );
  }

  PipelinePatch copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<String?>? name,
    Wrapped<Object?>? trigger,
    Wrapped<List<Object>?>? worker,
    Wrapped<List<Object>?>? edges,
  }) {
    return PipelinePatch(
      id: (id != null ? id.value : this.id),
      name: (name != null ? name.value : this.name),
      trigger: (trigger != null ? trigger.value : this.trigger),
      worker: (worker != null ? worker.value : this.worker),
      edges: (edges != null ? edges.value : this.edges),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class PipelineResponse {
  const PipelineResponse({
    required this.id,
    this.status,
    required this.name,
    this.trigger,
    this.worker,
    this.edges,
  });

  factory PipelineResponse.fromJson(Map<String, dynamic> json) =>
      _$PipelineResponseFromJson(json);

  static const toJsonFactory = _$PipelineResponseToJson;
  Map<String, dynamic> toJson() => _$PipelineResponseToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'status')
  final String? status;
  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'trigger')
  final Object? trigger;
  @JsonKey(name: 'worker', defaultValue: <Object>[])
  final List<Object>? worker;
  @JsonKey(name: 'edges', defaultValue: <Edge>[])
  final List<Edge>? edges;
  static const fromJsonFactory = _$PipelineResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PipelineResponse &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.status, status) ||
                const DeepCollectionEquality().equals(other.status, status)) &&
            (identical(other.name, name) ||
                const DeepCollectionEquality().equals(other.name, name)) &&
            (identical(other.trigger, trigger) ||
                const DeepCollectionEquality().equals(
                  other.trigger,
                  trigger,
                )) &&
            (identical(other.worker, worker) ||
                const DeepCollectionEquality().equals(other.worker, worker)) &&
            (identical(other.edges, edges) ||
                const DeepCollectionEquality().equals(other.edges, edges)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(status) ^
      const DeepCollectionEquality().hash(name) ^
      const DeepCollectionEquality().hash(trigger) ^
      const DeepCollectionEquality().hash(worker) ^
      const DeepCollectionEquality().hash(edges) ^
      runtimeType.hashCode;
}

extension $PipelineResponseExtension on PipelineResponse {
  PipelineResponse copyWith({
    String? id,
    String? status,
    String? name,
    Object? trigger,
    List<Object>? worker,
    List<Edge>? edges,
  }) {
    return PipelineResponse(
      id: id ?? this.id,
      status: status ?? this.status,
      name: name ?? this.name,
      trigger: trigger ?? this.trigger,
      worker: worker ?? this.worker,
      edges: edges ?? this.edges,
    );
  }

  PipelineResponse copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<String?>? status,
    Wrapped<String>? name,
    Wrapped<Object?>? trigger,
    Wrapped<List<Object>?>? worker,
    Wrapped<List<Edge>?>? edges,
  }) {
    return PipelineResponse(
      id: (id != null ? id.value : this.id),
      status: (status != null ? status.value : this.status),
      name: (name != null ? name.value : this.name),
      trigger: (trigger != null ? trigger.value : this.trigger),
      worker: (worker != null ? worker.value : this.worker),
      edges: (edges != null ? edges.value : this.edges),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class PortDefinition {
  const PortDefinition({
    required this.id,
    required this.label,
    required this.dataType,
    this.description,
    this.required,
    this.multiple,
  });

  factory PortDefinition.fromJson(Map<String, dynamic> json) =>
      _$PortDefinitionFromJson(json);

  static const toJsonFactory = _$PortDefinitionToJson;
  Map<String, dynamic> toJson() => _$PortDefinitionToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'label')
  final String label;
  @JsonKey(
    name: 'data_type',
    toJson: inputOutputTypeToJson,
    fromJson: inputOutputTypeFromJson,
  )
  final enums.InputOutputType dataType;
  @JsonKey(name: 'description')
  final String? description;
  @JsonKey(name: 'required', defaultValue: true)
  final bool? required;
  @JsonKey(name: 'multiple', defaultValue: false)
  final bool? multiple;
  static const fromJsonFactory = _$PortDefinitionFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PortDefinition &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.label, label) ||
                const DeepCollectionEquality().equals(other.label, label)) &&
            (identical(other.dataType, dataType) ||
                const DeepCollectionEquality().equals(
                  other.dataType,
                  dataType,
                )) &&
            (identical(other.description, description) ||
                const DeepCollectionEquality().equals(
                  other.description,
                  description,
                )) &&
            (identical(other.required, required) ||
                const DeepCollectionEquality().equals(
                  other.required,
                  required,
                )) &&
            (identical(other.multiple, multiple) ||
                const DeepCollectionEquality().equals(
                  other.multiple,
                  multiple,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(label) ^
      const DeepCollectionEquality().hash(dataType) ^
      const DeepCollectionEquality().hash(description) ^
      const DeepCollectionEquality().hash(required) ^
      const DeepCollectionEquality().hash(multiple) ^
      runtimeType.hashCode;
}

extension $PortDefinitionExtension on PortDefinition {
  PortDefinition copyWith({
    String? id,
    String? label,
    enums.InputOutputType? dataType,
    String? description,
    bool? required,
    bool? multiple,
  }) {
    return PortDefinition(
      id: id ?? this.id,
      label: label ?? this.label,
      dataType: dataType ?? this.dataType,
      description: description ?? this.description,
      required: required ?? this.required,
      multiple: multiple ?? this.multiple,
    );
  }

  PortDefinition copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<String>? label,
    Wrapped<enums.InputOutputType>? dataType,
    Wrapped<String?>? description,
    Wrapped<bool?>? required,
    Wrapped<bool?>? multiple,
  }) {
    return PortDefinition(
      id: (id != null ? id.value : this.id),
      label: (label != null ? label.value : this.label),
      dataType: (dataType != null ? dataType.value : this.dataType),
      description: (description != null ? description.value : this.description),
      required: (required != null ? required.value : this.required),
      multiple: (multiple != null ? multiple.value : this.multiple),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class TriggerConfig {
  const TriggerConfig({this.gridPosition});

  factory TriggerConfig.fromJson(Map<String, dynamic> json) =>
      _$TriggerConfigFromJson(json);

  static const toJsonFactory = _$TriggerConfigToJson;
  Map<String, dynamic> toJson() => _$TriggerConfigToJson(this);

  @JsonKey(name: 'gridPosition')
  final GridPosition? gridPosition;
  static const fromJsonFactory = _$TriggerConfigFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is TriggerConfig &&
            (identical(other.gridPosition, gridPosition) ||
                const DeepCollectionEquality().equals(
                  other.gridPosition,
                  gridPosition,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(gridPosition) ^ runtimeType.hashCode;
}

extension $TriggerConfigExtension on TriggerConfig {
  TriggerConfig copyWith({GridPosition? gridPosition}) {
    return TriggerConfig(gridPosition: gridPosition ?? this.gridPosition);
  }

  TriggerConfig copyWithWrapped({Wrapped<GridPosition?>? gridPosition}) {
    return TriggerConfig(
      gridPosition: (gridPosition != null
          ? gridPosition.value
          : this.gridPosition),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class TriggerConfigResponse {
  const TriggerConfigResponse({required this.id, this.gridPosition});

  factory TriggerConfigResponse.fromJson(Map<String, dynamic> json) =>
      _$TriggerConfigResponseFromJson(json);

  static const toJsonFactory = _$TriggerConfigResponseToJson;
  Map<String, dynamic> toJson() => _$TriggerConfigResponseToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'gridPosition')
  final GridPosition? gridPosition;
  static const fromJsonFactory = _$TriggerConfigResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is TriggerConfigResponse &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.gridPosition, gridPosition) ||
                const DeepCollectionEquality().equals(
                  other.gridPosition,
                  gridPosition,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(gridPosition) ^
      runtimeType.hashCode;
}

extension $TriggerConfigResponseExtension on TriggerConfigResponse {
  TriggerConfigResponse copyWith({String? id, GridPosition? gridPosition}) {
    return TriggerConfigResponse(
      id: id ?? this.id,
      gridPosition: gridPosition ?? this.gridPosition,
    );
  }

  TriggerConfigResponse copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<GridPosition?>? gridPosition,
  }) {
    return TriggerConfigResponse(
      id: (id != null ? id.value : this.id),
      gridPosition: (gridPosition != null
          ? gridPosition.value
          : this.gridPosition),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class ValidationError {
  const ValidationError({
    required this.loc,
    required this.msg,
    required this.type,
  });

  factory ValidationError.fromJson(Map<String, dynamic> json) =>
      _$ValidationErrorFromJson(json);

  static const toJsonFactory = _$ValidationErrorToJson;
  Map<String, dynamic> toJson() => _$ValidationErrorToJson(this);

  @JsonKey(name: 'loc', defaultValue: <Object>[])
  final List<Object> loc;
  @JsonKey(name: 'msg')
  final String msg;
  @JsonKey(name: 'type')
  final String type;
  static const fromJsonFactory = _$ValidationErrorFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is ValidationError &&
            (identical(other.loc, loc) ||
                const DeepCollectionEquality().equals(other.loc, loc)) &&
            (identical(other.msg, msg) ||
                const DeepCollectionEquality().equals(other.msg, msg)) &&
            (identical(other.type, type) ||
                const DeepCollectionEquality().equals(other.type, type)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(loc) ^
      const DeepCollectionEquality().hash(msg) ^
      const DeepCollectionEquality().hash(type) ^
      runtimeType.hashCode;
}

extension $ValidationErrorExtension on ValidationError {
  ValidationError copyWith({List<Object>? loc, String? msg, String? type}) {
    return ValidationError(
      loc: loc ?? this.loc,
      msg: msg ?? this.msg,
      type: type ?? this.type,
    );
  }

  ValidationError copyWithWrapped({
    Wrapped<List<Object>>? loc,
    Wrapped<String>? msg,
    Wrapped<String>? type,
  }) {
    return ValidationError(
      loc: (loc != null ? loc.value : this.loc),
      msg: (msg != null ? msg.value : this.msg),
      type: (type != null ? type.value : this.type),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class WorkerConfig {
  const WorkerConfig({
    this.gridPosition,
    this.numberOfInstances,
    this.instanceHealth,
  });

  factory WorkerConfig.fromJson(Map<String, dynamic> json) =>
      _$WorkerConfigFromJson(json);

  static const toJsonFactory = _$WorkerConfigToJson;
  Map<String, dynamic> toJson() => _$WorkerConfigToJson(this);

  @JsonKey(name: 'gridPosition')
  final GridPosition? gridPosition;
  @JsonKey(name: 'numberOfInstances')
  final int? numberOfInstances;
  @JsonKey(name: 'instanceHealth')
  final List<InstanceHealth>? instanceHealth;
  static const fromJsonFactory = _$WorkerConfigFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is WorkerConfig &&
            (identical(other.gridPosition, gridPosition) ||
                const DeepCollectionEquality().equals(
                  other.gridPosition,
                  gridPosition,
                )) &&
            (identical(other.numberOfInstances, numberOfInstances) ||
                const DeepCollectionEquality().equals(
                  other.numberOfInstances,
                  numberOfInstances,
                )) &&
            (identical(other.instanceHealth, instanceHealth) ||
                const DeepCollectionEquality().equals(
                  other.instanceHealth,
                  instanceHealth,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(gridPosition) ^
      const DeepCollectionEquality().hash(numberOfInstances) ^
      const DeepCollectionEquality().hash(instanceHealth) ^
      runtimeType.hashCode;
}

extension $WorkerConfigExtension on WorkerConfig {
  WorkerConfig copyWith({
    GridPosition? gridPosition,
    int? numberOfInstances,
    List<InstanceHealth>? instanceHealth,
  }) {
    return WorkerConfig(
      gridPosition: gridPosition ?? this.gridPosition,
      numberOfInstances: numberOfInstances ?? this.numberOfInstances,
      instanceHealth: instanceHealth ?? this.instanceHealth,
    );
  }

  WorkerConfig copyWithWrapped({
    Wrapped<GridPosition?>? gridPosition,
    Wrapped<int?>? numberOfInstances,
    Wrapped<List<InstanceHealth>?>? instanceHealth,
  }) {
    return WorkerConfig(
      gridPosition: (gridPosition != null
          ? gridPosition.value
          : this.gridPosition),
      numberOfInstances: (numberOfInstances != null
          ? numberOfInstances.value
          : this.numberOfInstances),
      instanceHealth: (instanceHealth != null
          ? instanceHealth.value
          : this.instanceHealth),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class WorkerConfigResponse {
  const WorkerConfigResponse({
    required this.id,
    this.gridPosition,
    this.numberOfInstances,
    this.instanceHealth,
  });

  factory WorkerConfigResponse.fromJson(Map<String, dynamic> json) =>
      _$WorkerConfigResponseFromJson(json);

  static const toJsonFactory = _$WorkerConfigResponseToJson;
  Map<String, dynamic> toJson() => _$WorkerConfigResponseToJson(this);

  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'gridPosition')
  final GridPosition? gridPosition;
  @JsonKey(name: 'numberOfInstances')
  final int? numberOfInstances;
  @JsonKey(name: 'instanceHealth')
  final List<InstanceHealthResponse>? instanceHealth;
  static const fromJsonFactory = _$WorkerConfigResponseFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is WorkerConfigResponse &&
            (identical(other.id, id) ||
                const DeepCollectionEquality().equals(other.id, id)) &&
            (identical(other.gridPosition, gridPosition) ||
                const DeepCollectionEquality().equals(
                  other.gridPosition,
                  gridPosition,
                )) &&
            (identical(other.numberOfInstances, numberOfInstances) ||
                const DeepCollectionEquality().equals(
                  other.numberOfInstances,
                  numberOfInstances,
                )) &&
            (identical(other.instanceHealth, instanceHealth) ||
                const DeepCollectionEquality().equals(
                  other.instanceHealth,
                  instanceHealth,
                )));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(id) ^
      const DeepCollectionEquality().hash(gridPosition) ^
      const DeepCollectionEquality().hash(numberOfInstances) ^
      const DeepCollectionEquality().hash(instanceHealth) ^
      runtimeType.hashCode;
}

extension $WorkerConfigResponseExtension on WorkerConfigResponse {
  WorkerConfigResponse copyWith({
    String? id,
    GridPosition? gridPosition,
    int? numberOfInstances,
    List<InstanceHealthResponse>? instanceHealth,
  }) {
    return WorkerConfigResponse(
      id: id ?? this.id,
      gridPosition: gridPosition ?? this.gridPosition,
      numberOfInstances: numberOfInstances ?? this.numberOfInstances,
      instanceHealth: instanceHealth ?? this.instanceHealth,
    );
  }

  WorkerConfigResponse copyWithWrapped({
    Wrapped<String>? id,
    Wrapped<GridPosition?>? gridPosition,
    Wrapped<int?>? numberOfInstances,
    Wrapped<List<InstanceHealthResponse>?>? instanceHealth,
  }) {
    return WorkerConfigResponse(
      id: (id != null ? id.value : this.id),
      gridPosition: (gridPosition != null
          ? gridPosition.value
          : this.gridPosition),
      numberOfInstances: (numberOfInstances != null
          ? numberOfInstances.value
          : this.numberOfInstances),
      instanceHealth: (instanceHealth != null
          ? instanceHealth.value
          : this.instanceHealth),
    );
  }
}

@JsonSerializable(explicitToJson: true)
class WorkerResult {
  const WorkerResult({
    required this.runId,
    required this.pipelineId,
    required this.nodeId,
    required this.status,
    this.outputPayload,
    this.rawData,
    this.error,
  });

  factory WorkerResult.fromJson(Map<String, dynamic> json) =>
      _$WorkerResultFromJson(json);

  static const toJsonFactory = _$WorkerResultToJson;
  Map<String, dynamic> toJson() => _$WorkerResultToJson(this);

  @JsonKey(name: 'run_id')
  final String runId;
  @JsonKey(name: 'pipeline_id')
  final String pipelineId;
  @JsonKey(name: 'node_id')
  final String nodeId;
  @JsonKey(
    name: 'status',
    toJson: nodeStatusToJson,
    fromJson: nodeStatusFromJson,
  )
  final enums.NodeStatus status;
  @JsonKey(name: 'output_payload')
  final JobPayload? outputPayload;
  @JsonKey(name: 'raw_data')
  final Object? rawData;
  @JsonKey(name: 'error')
  final String? error;
  static const fromJsonFactory = _$WorkerResultFromJson;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is WorkerResult &&
            (identical(other.runId, runId) ||
                const DeepCollectionEquality().equals(other.runId, runId)) &&
            (identical(other.pipelineId, pipelineId) ||
                const DeepCollectionEquality().equals(
                  other.pipelineId,
                  pipelineId,
                )) &&
            (identical(other.nodeId, nodeId) ||
                const DeepCollectionEquality().equals(other.nodeId, nodeId)) &&
            (identical(other.status, status) ||
                const DeepCollectionEquality().equals(other.status, status)) &&
            (identical(other.outputPayload, outputPayload) ||
                const DeepCollectionEquality().equals(
                  other.outputPayload,
                  outputPayload,
                )) &&
            (identical(other.rawData, rawData) ||
                const DeepCollectionEquality().equals(
                  other.rawData,
                  rawData,
                )) &&
            (identical(other.error, error) ||
                const DeepCollectionEquality().equals(other.error, error)));
  }

  @override
  String toString() => jsonEncode(this);

  @override
  int get hashCode =>
      const DeepCollectionEquality().hash(runId) ^
      const DeepCollectionEquality().hash(pipelineId) ^
      const DeepCollectionEquality().hash(nodeId) ^
      const DeepCollectionEquality().hash(status) ^
      const DeepCollectionEquality().hash(outputPayload) ^
      const DeepCollectionEquality().hash(rawData) ^
      const DeepCollectionEquality().hash(error) ^
      runtimeType.hashCode;
}

extension $WorkerResultExtension on WorkerResult {
  WorkerResult copyWith({
    String? runId,
    String? pipelineId,
    String? nodeId,
    enums.NodeStatus? status,
    JobPayload? outputPayload,
    Object? rawData,
    String? error,
  }) {
    return WorkerResult(
      runId: runId ?? this.runId,
      pipelineId: pipelineId ?? this.pipelineId,
      nodeId: nodeId ?? this.nodeId,
      status: status ?? this.status,
      outputPayload: outputPayload ?? this.outputPayload,
      rawData: rawData ?? this.rawData,
      error: error ?? this.error,
    );
  }

  WorkerResult copyWithWrapped({
    Wrapped<String>? runId,
    Wrapped<String>? pipelineId,
    Wrapped<String>? nodeId,
    Wrapped<enums.NodeStatus>? status,
    Wrapped<JobPayload?>? outputPayload,
    Wrapped<Object?>? rawData,
    Wrapped<String?>? error,
  }) {
    return WorkerResult(
      runId: (runId != null ? runId.value : this.runId),
      pipelineId: (pipelineId != null ? pipelineId.value : this.pipelineId),
      nodeId: (nodeId != null ? nodeId.value : this.nodeId),
      status: (status != null ? status.value : this.status),
      outputPayload: (outputPayload != null
          ? outputPayload.value
          : this.outputPayload),
      rawData: (rawData != null ? rawData.value : this.rawData),
      error: (error != null ? error.value : this.error),
    );
  }
}

String? fieldTypeNullableToJson(enums.FieldType? fieldType) {
  return fieldType?.value;
}

String? fieldTypeToJson(enums.FieldType fieldType) {
  return fieldType.value;
}

enums.FieldType fieldTypeFromJson(
  Object? fieldType, [
  enums.FieldType? defaultValue,
]) {
  return enums.FieldType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            fieldType?.toString().toLowerCase(),
      ) ??
      defaultValue ??
      enums.FieldType.swaggerGeneratedUnknown;
}

enums.FieldType? fieldTypeNullableFromJson(
  Object? fieldType, [
  enums.FieldType? defaultValue,
]) {
  if (fieldType == null) {
    return null;
  }
  return enums.FieldType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            fieldType.toString().toLowerCase(),
      ) ??
      defaultValue;
}

String fieldTypeExplodedListToJson(List<enums.FieldType>? fieldType) {
  return fieldType?.map((e) => e.value!).join(',') ?? '';
}

List<String> fieldTypeListToJson(List<enums.FieldType>? fieldType) {
  if (fieldType == null) {
    return [];
  }

  return fieldType.map((e) => e.value!).toList();
}

List<enums.FieldType> fieldTypeListFromJson(
  List? fieldType, [
  List<enums.FieldType>? defaultValue,
]) {
  if (fieldType == null) {
    return defaultValue ?? [];
  }

  return fieldType.map((e) => fieldTypeFromJson(e.toString())).toList();
}

List<enums.FieldType>? fieldTypeNullableListFromJson(
  List? fieldType, [
  List<enums.FieldType>? defaultValue,
]) {
  if (fieldType == null) {
    return defaultValue;
  }

  return fieldType.map((e) => fieldTypeFromJson(e.toString())).toList();
}

String? inputOutputTypeNullableToJson(enums.InputOutputType? inputOutputType) {
  return inputOutputType?.value;
}

String? inputOutputTypeToJson(enums.InputOutputType inputOutputType) {
  return inputOutputType.value;
}

enums.InputOutputType inputOutputTypeFromJson(
  Object? inputOutputType, [
  enums.InputOutputType? defaultValue,
]) {
  return enums.InputOutputType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            inputOutputType?.toString().toLowerCase(),
      ) ??
      defaultValue ??
      enums.InputOutputType.swaggerGeneratedUnknown;
}

enums.InputOutputType? inputOutputTypeNullableFromJson(
  Object? inputOutputType, [
  enums.InputOutputType? defaultValue,
]) {
  if (inputOutputType == null) {
    return null;
  }
  return enums.InputOutputType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            inputOutputType.toString().toLowerCase(),
      ) ??
      defaultValue;
}

String inputOutputTypeExplodedListToJson(
  List<enums.InputOutputType>? inputOutputType,
) {
  return inputOutputType?.map((e) => e.value!).join(',') ?? '';
}

List<String> inputOutputTypeListToJson(
  List<enums.InputOutputType>? inputOutputType,
) {
  if (inputOutputType == null) {
    return [];
  }

  return inputOutputType.map((e) => e.value!).toList();
}

List<enums.InputOutputType> inputOutputTypeListFromJson(
  List? inputOutputType, [
  List<enums.InputOutputType>? defaultValue,
]) {
  if (inputOutputType == null) {
    return defaultValue ?? [];
  }

  return inputOutputType
      .map((e) => inputOutputTypeFromJson(e.toString()))
      .toList();
}

List<enums.InputOutputType>? inputOutputTypeNullableListFromJson(
  List? inputOutputType, [
  List<enums.InputOutputType>? defaultValue,
]) {
  if (inputOutputType == null) {
    return defaultValue;
  }

  return inputOutputType
      .map((e) => inputOutputTypeFromJson(e.toString()))
      .toList();
}

String? nodeStatusNullableToJson(enums.NodeStatus? nodeStatus) {
  return nodeStatus?.value;
}

String? nodeStatusToJson(enums.NodeStatus nodeStatus) {
  return nodeStatus.value;
}

enums.NodeStatus nodeStatusFromJson(
  Object? nodeStatus, [
  enums.NodeStatus? defaultValue,
]) {
  return enums.NodeStatus.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            nodeStatus?.toString().toLowerCase(),
      ) ??
      defaultValue ??
      enums.NodeStatus.swaggerGeneratedUnknown;
}

enums.NodeStatus? nodeStatusNullableFromJson(
  Object? nodeStatus, [
  enums.NodeStatus? defaultValue,
]) {
  if (nodeStatus == null) {
    return null;
  }
  return enums.NodeStatus.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            nodeStatus.toString().toLowerCase(),
      ) ??
      defaultValue;
}

String nodeStatusExplodedListToJson(List<enums.NodeStatus>? nodeStatus) {
  return nodeStatus?.map((e) => e.value!).join(',') ?? '';
}

List<String> nodeStatusListToJson(List<enums.NodeStatus>? nodeStatus) {
  if (nodeStatus == null) {
    return [];
  }

  return nodeStatus.map((e) => e.value!).toList();
}

List<enums.NodeStatus> nodeStatusListFromJson(
  List? nodeStatus, [
  List<enums.NodeStatus>? defaultValue,
]) {
  if (nodeStatus == null) {
    return defaultValue ?? [];
  }

  return nodeStatus.map((e) => nodeStatusFromJson(e.toString())).toList();
}

List<enums.NodeStatus>? nodeStatusNullableListFromJson(
  List? nodeStatus, [
  List<enums.NodeStatus>? defaultValue,
]) {
  if (nodeStatus == null) {
    return defaultValue;
  }

  return nodeStatus.map((e) => nodeStatusFromJson(e.toString())).toList();
}

String? nodeTypeNullableToJson(enums.NodeType? nodeType) {
  return nodeType?.value;
}

String? nodeTypeToJson(enums.NodeType nodeType) {
  return nodeType.value;
}

enums.NodeType nodeTypeFromJson(
  Object? nodeType, [
  enums.NodeType? defaultValue,
]) {
  return enums.NodeType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            nodeType?.toString().toLowerCase(),
      ) ??
      defaultValue ??
      enums.NodeType.swaggerGeneratedUnknown;
}

enums.NodeType? nodeTypeNullableFromJson(
  Object? nodeType, [
  enums.NodeType? defaultValue,
]) {
  if (nodeType == null) {
    return null;
  }
  return enums.NodeType.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            nodeType.toString().toLowerCase(),
      ) ??
      defaultValue;
}

String nodeTypeExplodedListToJson(List<enums.NodeType>? nodeType) {
  return nodeType?.map((e) => e.value!).join(',') ?? '';
}

List<String> nodeTypeListToJson(List<enums.NodeType>? nodeType) {
  if (nodeType == null) {
    return [];
  }

  return nodeType.map((e) => e.value!).toList();
}

List<enums.NodeType> nodeTypeListFromJson(
  List? nodeType, [
  List<enums.NodeType>? defaultValue,
]) {
  if (nodeType == null) {
    return defaultValue ?? [];
  }

  return nodeType.map((e) => nodeTypeFromJson(e.toString())).toList();
}

List<enums.NodeType>? nodeTypeNullableListFromJson(
  List? nodeType, [
  List<enums.NodeType>? defaultValue,
]) {
  if (nodeType == null) {
    return defaultValue;
  }

  return nodeType.map((e) => nodeTypeFromJson(e.toString())).toList();
}

String? workerCategoryNullableToJson(enums.WorkerCategory? workerCategory) {
  return workerCategory?.value;
}

String? workerCategoryToJson(enums.WorkerCategory workerCategory) {
  return workerCategory.value;
}

enums.WorkerCategory workerCategoryFromJson(
  Object? workerCategory, [
  enums.WorkerCategory? defaultValue,
]) {
  return enums.WorkerCategory.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            workerCategory?.toString().toLowerCase(),
      ) ??
      defaultValue ??
      enums.WorkerCategory.swaggerGeneratedUnknown;
}

enums.WorkerCategory? workerCategoryNullableFromJson(
  Object? workerCategory, [
  enums.WorkerCategory? defaultValue,
]) {
  if (workerCategory == null) {
    return null;
  }
  return enums.WorkerCategory.values.firstWhereOrNull(
        (e) =>
            e.value.toString().toLowerCase() ==
            workerCategory.toString().toLowerCase(),
      ) ??
      defaultValue;
}

String workerCategoryExplodedListToJson(
  List<enums.WorkerCategory>? workerCategory,
) {
  return workerCategory?.map((e) => e.value!).join(',') ?? '';
}

List<String> workerCategoryListToJson(
  List<enums.WorkerCategory>? workerCategory,
) {
  if (workerCategory == null) {
    return [];
  }

  return workerCategory.map((e) => e.value!).toList();
}

List<enums.WorkerCategory> workerCategoryListFromJson(
  List? workerCategory, [
  List<enums.WorkerCategory>? defaultValue,
]) {
  if (workerCategory == null) {
    return defaultValue ?? [];
  }

  return workerCategory
      .map((e) => workerCategoryFromJson(e.toString()))
      .toList();
}

List<enums.WorkerCategory>? workerCategoryNullableListFromJson(
  List? workerCategory, [
  List<enums.WorkerCategory>? defaultValue,
]) {
  if (workerCategory == null) {
    return defaultValue;
  }

  return workerCategory
      .map((e) => workerCategoryFromJson(e.toString()))
      .toList();
}

// ignore: unused_element
String? _dateToJson(DateTime? date) {
  if (date == null) {
    return null;
  }

  final year = date.year.toString();
  final month = date.month < 10 ? '0${date.month}' : date.month.toString();
  final day = date.day < 10 ? '0${date.day}' : date.day.toString();

  return '$year-$month-$day';
}

class Wrapped<T> {
  final T value;
  const Wrapped.value(this.value);
}
