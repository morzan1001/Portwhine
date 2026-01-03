// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'portwhine.models.swagger.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DeleteResponse _$DeleteResponseFromJson(Map<String, dynamic> json) =>
    DeleteResponse(detail: json['detail'] as String?);

Map<String, dynamic> _$DeleteResponseToJson(DeleteResponse instance) =>
    <String, dynamic>{'detail': instance.detail};

Edge _$EdgeFromJson(Map<String, dynamic> json) => Edge(
  source: json['source'] as String,
  target: json['target'] as String,
  sourcePort: json['source_port'] as String?,
  targetPort: json['target_port'] as String?,
);

Map<String, dynamic> _$EdgeToJson(Edge instance) => <String, dynamic>{
  'source': instance.source,
  'target': instance.target,
  'source_port': instance.sourcePort,
  'target_port': instance.targetPort,
};

FieldDefinition _$FieldDefinitionFromJson(Map<String, dynamic> json) =>
    FieldDefinition(
      name: json['name'] as String,
      label: json['label'] as String,
      type: fieldTypeFromJson(json['type']),
      description: json['description'] as String?,
      required: json['required'] as bool? ?? false,
      $default: json['default'],
      options:
          (json['options'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          [],
      placeholder: json['placeholder'] as String?,
      validationPattern: json['validation_pattern'] as String?,
      minValue: (json['min_value'] as num?)?.toDouble(),
      maxValue: (json['max_value'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$FieldDefinitionToJson(FieldDefinition instance) =>
    <String, dynamic>{
      'name': instance.name,
      'label': instance.label,
      'type': fieldTypeToJson(instance.type),
      'description': instance.description,
      'required': instance.required,
      'default': instance.$default,
      'options': instance.options,
      'placeholder': instance.placeholder,
      'validation_pattern': instance.validationPattern,
      'min_value': instance.minValue,
      'max_value': instance.maxValue,
    };

GridPosition _$GridPositionFromJson(Map<String, dynamic> json) => GridPosition(
  x: (json['x'] as num?)?.toDouble(),
  y: (json['y'] as num?)?.toDouble(),
);

Map<String, dynamic> _$GridPositionToJson(GridPosition instance) =>
    <String, dynamic>{'x': instance.x, 'y': instance.y};

HTTPValidationError _$HTTPValidationErrorFromJson(Map<String, dynamic> json) =>
    HTTPValidationError(
      detail:
          (json['detail'] as List<dynamic>?)
              ?.map((e) => ValidationError.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$HTTPValidationErrorToJson(
  HTTPValidationError instance,
) => <String, dynamic>{
  'detail': instance.detail?.map((e) => e.toJson()).toList(),
};

HttpTarget _$HttpTargetFromJson(Map<String, dynamic> json) => HttpTarget(
  url: json['url'] as String,
  method: json['method'] as String?,
  headers: json['headers'] as Map<String, dynamic>?,
);

Map<String, dynamic> _$HttpTargetToJson(HttpTarget instance) =>
    <String, dynamic>{
      'url': instance.url,
      'method': instance.method,
      'headers': instance.headers,
    };

InstanceHealth _$InstanceHealthFromJson(Map<String, dynamic> json) =>
    InstanceHealth(
      instanceNumber: (json['instance_number'] as num).toInt(),
      containerId: json['container_id'] as String?,
      containerName: json['container_name'] as String?,
      status: InstanceHealth.nodeStatusStatusNullableFromJson(json['status']),
      startedAt: json['started_at'] == null
          ? null
          : DateTime.parse(json['started_at'] as String),
      finishedAt: json['finished_at'] == null
          ? null
          : DateTime.parse(json['finished_at'] as String),
      exitCode: (json['exit_code'] as num?)?.toInt(),
      errorMessage: json['error_message'] as String?,
      jobsProcessed: (json['jobs_processed'] as num?)?.toInt(),
    );

Map<String, dynamic> _$InstanceHealthToJson(InstanceHealth instance) =>
    <String, dynamic>{
      'instance_number': instance.instanceNumber,
      'container_id': instance.containerId,
      'container_name': instance.containerName,
      'status': nodeStatusNullableToJson(instance.status),
      'started_at': instance.startedAt?.toIso8601String(),
      'finished_at': instance.finishedAt?.toIso8601String(),
      'exit_code': instance.exitCode,
      'error_message': instance.errorMessage,
      'jobs_processed': instance.jobsProcessed,
    };

InstanceHealthResponse _$InstanceHealthResponseFromJson(
  Map<String, dynamic> json,
) => InstanceHealthResponse(
  instanceNumber: (json['instance_number'] as num).toInt(),
  containerId: json['container_id'] as String?,
  containerName: json['container_name'] as String?,
  status: json['status'] as String?,
  startedAt: json['started_at'] as String?,
  finishedAt: json['finished_at'] as String?,
  exitCode: (json['exit_code'] as num?)?.toInt(),
  errorMessage: json['error_message'] as String?,
  jobsProcessed: (json['jobs_processed'] as num?)?.toInt(),
);

Map<String, dynamic> _$InstanceHealthResponseToJson(
  InstanceHealthResponse instance,
) => <String, dynamic>{
  'instance_number': instance.instanceNumber,
  'container_id': instance.containerId,
  'container_name': instance.containerName,
  'status': instance.status,
  'started_at': instance.startedAt,
  'finished_at': instance.finishedAt,
  'exit_code': instance.exitCode,
  'error_message': instance.errorMessage,
  'jobs_processed': instance.jobsProcessed,
};

IpTarget _$IpTargetFromJson(Map<String, dynamic> json) =>
    IpTarget(ip: json['ip'], port: (json['port'] as num?)?.toInt());

Map<String, dynamic> _$IpTargetToJson(IpTarget instance) => <String, dynamic>{
  'ip': instance.ip,
  'port': instance.port,
};

JobPayload _$JobPayloadFromJson(Map<String, dynamic> json) => JobPayload(
  http:
      (json['http'] as List<dynamic>?)
          ?.map((e) => HttpTarget.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
  ip:
      (json['ip'] as List<dynamic>?)
          ?.map((e) => IpTarget.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
);

Map<String, dynamic> _$JobPayloadToJson(JobPayload instance) =>
    <String, dynamic>{
      'http': instance.http?.map((e) => e.toJson()).toList(),
      'ip': instance.ip?.map((e) => e.toJson()).toList(),
    };

MessageResponse _$MessageResponseFromJson(Map<String, dynamic> json) =>
    MessageResponse(detail: json['detail'] as String);

Map<String, dynamic> _$MessageResponseToJson(MessageResponse instance) =>
    <String, dynamic>{'detail': instance.detail};

NodeConfigExampleResponse _$NodeConfigExampleResponseFromJson(
  Map<String, dynamic> json,
) => NodeConfigExampleResponse(
  description: json['description'] as String,
  example: json['example'] as Map<String, dynamic>,
);

Map<String, dynamic> _$NodeConfigExampleResponseToJson(
  NodeConfigExampleResponse instance,
) => <String, dynamic>{
  'description': instance.description,
  'example': instance.example,
};

NodeDefinition _$NodeDefinitionFromJson(Map<String, dynamic> json) =>
    NodeDefinition(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      nodeType: nodeTypeFromJson(json['node_type']),
      category: workerCategoryNullableFromJson(json['category']),
      icon: json['icon'] as String?,
      color: json['color'] as String?,
      inputs:
          (json['inputs'] as List<dynamic>?)
              ?.map((e) => PortDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      outputs:
          (json['outputs'] as List<dynamic>?)
              ?.map((e) => PortDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      configFields:
          (json['config_fields'] as List<dynamic>?)
              ?.map((e) => FieldDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      imageName: json['image_name'] as String,
      supportsMultipleInstances:
          json['supports_multiple_instances'] as bool? ?? true,
      maxInstances: (json['max_instances'] as num?)?.toInt(),
    );

Map<String, dynamic> _$NodeDefinitionToJson(NodeDefinition instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'description': instance.description,
      'node_type': nodeTypeToJson(instance.nodeType),
      'category': workerCategoryNullableToJson(instance.category),
      'icon': instance.icon,
      'color': instance.color,
      'inputs': instance.inputs?.map((e) => e.toJson()).toList(),
      'outputs': instance.outputs?.map((e) => e.toJson()).toList(),
      'config_fields': instance.configFields?.map((e) => e.toJson()).toList(),
      'image_name': instance.imageName,
      'supports_multiple_instances': instance.supportsMultipleInstances,
      'max_instances': instance.maxInstances,
    };

Pipeline _$PipelineFromJson(Map<String, dynamic> json) => Pipeline(
  name: json['name'] as String,
  trigger: json['trigger'] == null
      ? null
      : TriggerConfig.fromJson(json['trigger'] as Map<String, dynamic>),
  worker: (json['worker'] as List<dynamic>?)
      ?.map((e) => WorkerConfig.fromJson(e as Map<String, dynamic>))
      .toList(),
  edges:
      (json['edges'] as List<dynamic>?)
          ?.map((e) => Edge.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
);

Map<String, dynamic> _$PipelineToJson(Pipeline instance) => <String, dynamic>{
  'name': instance.name,
  'trigger': instance.trigger?.toJson(),
  'worker': instance.worker?.map((e) => e.toJson()).toList(),
  'edges': instance.edges?.map((e) => e.toJson()).toList(),
};

PipelineListItem _$PipelineListItemFromJson(Map<String, dynamic> json) =>
    PipelineListItem(
      id: json['id'] as String,
      name: json['name'] as String,
      status: json['status'] as String?,
    );

Map<String, dynamic> _$PipelineListItemToJson(PipelineListItem instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'status': instance.status,
    };

PipelinePatch _$PipelinePatchFromJson(
  Map<String, dynamic> json,
) => PipelinePatch(
  id: json['id'] as String,
  name: json['name'] as String?,
  trigger: json['trigger'],
  worker: (json['worker'] as List<dynamic>?)?.map((e) => e as Object).toList(),
  edges: (json['edges'] as List<dynamic>?)?.map((e) => e as Object).toList(),
);

Map<String, dynamic> _$PipelinePatchToJson(PipelinePatch instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'trigger': instance.trigger,
      'worker': instance.worker,
      'edges': instance.edges,
    };

PipelineResponse _$PipelineResponseFromJson(Map<String, dynamic> json) =>
    PipelineResponse(
      id: json['id'] as String,
      status: json['status'] as String?,
      name: json['name'] as String,
      trigger: json['trigger'],
      worker:
          (json['worker'] as List<dynamic>?)
              ?.map((e) => e as Object)
              .toList() ??
          [],
      edges:
          (json['edges'] as List<dynamic>?)
              ?.map((e) => Edge.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$PipelineResponseToJson(PipelineResponse instance) =>
    <String, dynamic>{
      'id': instance.id,
      'status': instance.status,
      'name': instance.name,
      'trigger': instance.trigger,
      'worker': instance.worker,
      'edges': instance.edges?.map((e) => e.toJson()).toList(),
    };

PortDefinition _$PortDefinitionFromJson(Map<String, dynamic> json) =>
    PortDefinition(
      id: json['id'] as String,
      label: json['label'] as String,
      dataType: inputOutputTypeFromJson(json['data_type']),
      description: json['description'] as String?,
      required: json['required'] as bool? ?? true,
      multiple: json['multiple'] as bool? ?? false,
    );

Map<String, dynamic> _$PortDefinitionToJson(PortDefinition instance) =>
    <String, dynamic>{
      'id': instance.id,
      'label': instance.label,
      'data_type': inputOutputTypeToJson(instance.dataType),
      'description': instance.description,
      'required': instance.required,
      'multiple': instance.multiple,
    };

TriggerConfig _$TriggerConfigFromJson(Map<String, dynamic> json) =>
    TriggerConfig(
      gridPosition: json['gridPosition'] == null
          ? null
          : GridPosition.fromJson(json['gridPosition'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$TriggerConfigToJson(TriggerConfig instance) =>
    <String, dynamic>{'gridPosition': instance.gridPosition?.toJson()};

TriggerConfigResponse _$TriggerConfigResponseFromJson(
  Map<String, dynamic> json,
) => TriggerConfigResponse(
  id: json['id'] as String,
  gridPosition: json['gridPosition'] == null
      ? null
      : GridPosition.fromJson(json['gridPosition'] as Map<String, dynamic>),
);

Map<String, dynamic> _$TriggerConfigResponseToJson(
  TriggerConfigResponse instance,
) => <String, dynamic>{
  'id': instance.id,
  'gridPosition': instance.gridPosition?.toJson(),
};

ValidationError _$ValidationErrorFromJson(
  Map<String, dynamic> json,
) => ValidationError(
  loc: (json['loc'] as List<dynamic>?)?.map((e) => e as Object).toList() ?? [],
  msg: json['msg'] as String,
  type: json['type'] as String,
);

Map<String, dynamic> _$ValidationErrorToJson(ValidationError instance) =>
    <String, dynamic>{
      'loc': instance.loc,
      'msg': instance.msg,
      'type': instance.type,
    };

WorkerConfig _$WorkerConfigFromJson(Map<String, dynamic> json) => WorkerConfig(
  gridPosition: json['gridPosition'] == null
      ? null
      : GridPosition.fromJson(json['gridPosition'] as Map<String, dynamic>),
  numberOfInstances: (json['numberOfInstances'] as num?)?.toInt(),
  instanceHealth: (json['instanceHealth'] as List<dynamic>?)
      ?.map((e) => InstanceHealth.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$WorkerConfigToJson(
  WorkerConfig instance,
) => <String, dynamic>{
  'gridPosition': instance.gridPosition?.toJson(),
  'numberOfInstances': instance.numberOfInstances,
  'instanceHealth': instance.instanceHealth?.map((e) => e.toJson()).toList(),
};

WorkerConfigResponse _$WorkerConfigResponseFromJson(
  Map<String, dynamic> json,
) => WorkerConfigResponse(
  id: json['id'] as String,
  gridPosition: json['gridPosition'] == null
      ? null
      : GridPosition.fromJson(json['gridPosition'] as Map<String, dynamic>),
  numberOfInstances: (json['numberOfInstances'] as num?)?.toInt(),
  instanceHealth: (json['instanceHealth'] as List<dynamic>?)
      ?.map((e) => InstanceHealthResponse.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$WorkerConfigResponseToJson(
  WorkerConfigResponse instance,
) => <String, dynamic>{
  'id': instance.id,
  'gridPosition': instance.gridPosition?.toJson(),
  'numberOfInstances': instance.numberOfInstances,
  'instanceHealth': instance.instanceHealth?.map((e) => e.toJson()).toList(),
};

WorkerResult _$WorkerResultFromJson(Map<String, dynamic> json) => WorkerResult(
  runId: json['run_id'] as String,
  pipelineId: json['pipeline_id'] as String,
  nodeId: json['node_id'] as String,
  status: nodeStatusFromJson(json['status']),
  outputPayload: json['output_payload'] == null
      ? null
      : JobPayload.fromJson(json['output_payload'] as Map<String, dynamic>),
  rawData: json['raw_data'],
  error: json['error'] as String?,
);

Map<String, dynamic> _$WorkerResultToJson(WorkerResult instance) =>
    <String, dynamic>{
      'run_id': instance.runId,
      'pipeline_id': instance.pipelineId,
      'node_id': instance.nodeId,
      'status': nodeStatusToJson(instance.status),
      'output_payload': instance.outputPayload?.toJson(),
      'raw_data': instance.rawData,
      'error': instance.error,
    };
