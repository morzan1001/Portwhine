// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element_parameter

import 'package:json_annotation/json_annotation.dart';
import 'package:json_annotation/json_annotation.dart' as json;
import 'package:collection/collection.dart';
import 'dart:convert';

import 'portwhine.models.swagger.dart';
import 'package:chopper/chopper.dart';

import 'client_mapping.dart';
import 'dart:async';
import 'package:http/http.dart' as http;
import 'package:http/http.dart' show MultipartFile;
import 'package:chopper/chopper.dart' as chopper;
import 'portwhine.enums.swagger.dart' as enums;
import 'portwhine.metadata.swagger.dart';
export 'portwhine.enums.swagger.dart';
export 'portwhine.models.swagger.dart';

part 'portwhine.swagger.chopper.dart';

// **************************************************************************
// SwaggerChopperGenerator
// **************************************************************************

@ChopperApi()
abstract class Portwhine extends ChopperService {
  static Portwhine create({
    ChopperClient? client,
    http.Client? httpClient,
    Authenticator? authenticator,
    ErrorConverter? errorConverter,
    Converter? converter,
    Uri? baseUrl,
    List<Interceptor>? interceptors,
  }) {
    if (client != null) {
      return _$Portwhine(client);
    }

    final newClient = ChopperClient(
      services: [_$Portwhine()],
      converter: converter ?? $JsonSerializableConverter(),
      interceptors: interceptors ?? [],
      client: httpClient,
      authenticator: authenticator,
      errorConverter: errorConverter,
      baseUrl: baseUrl,
    );
    return _$Portwhine(newClient);
  }

  ///Get all trigger names
  Future<chopper.Response<List<String>>> apiV1TriggerGet() {
    return _apiV1TriggerGet();
  }

  ///Get all trigger names
  @GET(path: '/api/v1/trigger')
  Future<chopper.Response<List<String>>> _apiV1TriggerGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Retrieve a list of all available trigger names.',
      summary: 'Get all trigger names',
      operationId: 'get_triggers_api_v1_trigger_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Trigger"],
      deprecated: false,
    ),
  });

  ///Get the configuration of a specific trigger
  ///@param name
  Future<chopper.Response<NodeConfigExampleResponse>> apiV1TriggerNameGet({
    required String? name,
  }) {
    generatedMapping.putIfAbsent(
      NodeConfigExampleResponse,
      () => NodeConfigExampleResponse.fromJsonFactory,
    );

    return _apiV1TriggerNameGet(name: name);
  }

  ///Get the configuration of a specific trigger
  ///@param name
  @GET(path: '/api/v1/trigger/{name}')
  Future<chopper.Response<NodeConfigExampleResponse>> _apiV1TriggerNameGet({
    @Path('name') required String? name,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Retrieve the configuration of a specific trigger by name.',
      summary: 'Get the configuration of a specific trigger',
      operationId: 'get_trigger_config_api_v1_trigger__name__get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Trigger"],
      deprecated: false,
    ),
  });

  ///Create a new pipeline
  Future<chopper.Response<PipelineResponse>> apiV1PipelinePost({
    required Pipeline? body,
  }) {
    generatedMapping.putIfAbsent(
      PipelineResponse,
      () => PipelineResponse.fromJsonFactory,
    );

    return _apiV1PipelinePost(body: body);
  }

  ///Create a new pipeline
  @POST(path: '/api/v1/pipeline', optionalBody: true)
  Future<chopper.Response<PipelineResponse>> _apiV1PipelinePost({
    @Body() required Pipeline? body,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Create a new pipeline with the specified configurations.',
      summary: 'Create a new pipeline',
      operationId: 'create_pipeline_api_v1_pipeline_post',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Update a pipeline configuration
  Future<chopper.Response<PipelineResponse>> apiV1PipelinePatch({
    required PipelinePatch? body,
  }) {
    generatedMapping.putIfAbsent(
      PipelineResponse,
      () => PipelineResponse.fromJsonFactory,
    );

    return _apiV1PipelinePatch(body: body);
  }

  ///Update a pipeline configuration
  @PATCH(path: '/api/v1/pipeline', optionalBody: true)
  Future<chopper.Response<PipelineResponse>> _apiV1PipelinePatch({
    @Body() required PipelinePatch? body,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description:
          'Update the configuration of a specific pipeline by pipeline ID. None set fields will be left unchanged.',
      summary: 'Update a pipeline configuration',
      operationId: 'update_pipeline_api_v1_pipeline_patch',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Get a pipeline configuration
  ///@param pipeline_id
  Future<chopper.Response<PipelineResponse>> apiV1PipelinePipelineIdGet({
    required String? pipelineId,
  }) {
    generatedMapping.putIfAbsent(
      PipelineResponse,
      () => PipelineResponse.fromJsonFactory,
    );

    return _apiV1PipelinePipelineIdGet(pipelineId: pipelineId);
  }

  ///Get a pipeline configuration
  ///@param pipeline_id
  @GET(path: '/api/v1/pipeline/{pipeline_id}')
  Future<chopper.Response<PipelineResponse>> _apiV1PipelinePipelineIdGet({
    @Path('pipeline_id') required String? pipelineId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description:
          'Retrieve the configuration of a specific pipeline by pipeline ID.',
      summary: 'Get a pipeline configuration',
      operationId: 'get_pipeline_api_v1_pipeline__pipeline_id__get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Delete a pipeline configuration
  ///@param pipeline_id
  Future<chopper.Response<DeleteResponse>> apiV1PipelinePipelineIdDelete({
    required String? pipelineId,
  }) {
    generatedMapping.putIfAbsent(
      DeleteResponse,
      () => DeleteResponse.fromJsonFactory,
    );

    return _apiV1PipelinePipelineIdDelete(pipelineId: pipelineId);
  }

  ///Delete a pipeline configuration
  ///@param pipeline_id
  @DELETE(path: '/api/v1/pipeline/{pipeline_id}')
  Future<chopper.Response<DeleteResponse>> _apiV1PipelinePipelineIdDelete({
    @Path('pipeline_id') required String? pipelineId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Delete a specific pipeline by pipeline ID.',
      summary: 'Delete a pipeline configuration',
      operationId: 'delete_pipeline_api_v1_pipeline__pipeline_id__delete',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Get all pipeline configurations
  ///@param size
  ///@param page
  Future<chopper.Response<List<PipelineListItem>>> apiV1PipelinesGet({
    int? size,
    int? page,
  }) {
    generatedMapping.putIfAbsent(
      PipelineListItem,
      () => PipelineListItem.fromJsonFactory,
    );

    return _apiV1PipelinesGet(size: size, page: page);
  }

  ///Get all pipeline configurations
  ///@param size
  ///@param page
  @GET(path: '/api/v1/pipelines')
  Future<chopper.Response<List<PipelineListItem>>> _apiV1PipelinesGet({
    @Query('size') int? size,
    @Query('page') int? page,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Retrieve a list of all available pipelines.',
      summary: 'Get all pipeline configurations',
      operationId: 'get_all_pipelines_api_v1_pipelines_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Start a pipeline
  ///@param pipeline_id
  Future<chopper.Response<MessageResponse>> apiV1PipelineStartPipelineIdPost({
    required String? pipelineId,
  }) {
    generatedMapping.putIfAbsent(
      MessageResponse,
      () => MessageResponse.fromJsonFactory,
    );

    return _apiV1PipelineStartPipelineIdPost(pipelineId: pipelineId);
  }

  ///Start a pipeline
  ///@param pipeline_id
  @POST(path: '/api/v1/pipeline/start/{pipeline_id}', optionalBody: true)
  Future<chopper.Response<MessageResponse>> _apiV1PipelineStartPipelineIdPost({
    @Path('pipeline_id') required String? pipelineId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Start a specific pipeline by pipeline ID.',
      summary: 'Start a pipeline',
      operationId: 'start_pipeline_api_v1_pipeline_start__pipeline_id__post',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Stop a pipeline
  ///@param pipeline_id
  Future<chopper.Response<MessageResponse>> apiV1PipelineStopPipelineIdPost({
    required String? pipelineId,
  }) {
    generatedMapping.putIfAbsent(
      MessageResponse,
      () => MessageResponse.fromJsonFactory,
    );

    return _apiV1PipelineStopPipelineIdPost(pipelineId: pipelineId);
  }

  ///Stop a pipeline
  ///@param pipeline_id
  @POST(path: '/api/v1/pipeline/stop/{pipeline_id}', optionalBody: true)
  Future<chopper.Response<MessageResponse>> _apiV1PipelineStopPipelineIdPost({
    @Path('pipeline_id') required String? pipelineId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Stop a specific pipeline by pipeline ID.',
      summary: 'Stop a pipeline',
      operationId: 'stop_pipeline_api_v1_pipeline_stop__pipeline_id__post',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Cleanup all containers for a pipeline
  ///@param pipeline_id
  Future<chopper.Response<MessageResponse>> apiV1PipelineCleanupPipelineIdPost({
    required String? pipelineId,
  }) {
    generatedMapping.putIfAbsent(
      MessageResponse,
      () => MessageResponse.fromJsonFactory,
    );

    return _apiV1PipelineCleanupPipelineIdPost(pipelineId: pipelineId);
  }

  ///Cleanup all containers for a pipeline
  ///@param pipeline_id
  @POST(path: '/api/v1/pipeline/cleanup/{pipeline_id}', optionalBody: true)
  Future<chopper.Response<MessageResponse>>
  _apiV1PipelineCleanupPipelineIdPost({
    @Path('pipeline_id') required String? pipelineId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description:
          'Cleanup all containers for a specific pipeline by pipeline ID.',
      summary: 'Cleanup all containers for a pipeline',
      operationId:
          'cleanup_containers_api_v1_pipeline_cleanup__pipeline_id__post',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Pipelines"],
      deprecated: false,
    ),
  });

  ///Handle Job Result
  ///@param instance_name
  Future<chopper.Response> apiV1JobResultPost({
    String? instanceName,
    required WorkerResult? body,
  }) {
    return _apiV1JobResultPost(instanceName: instanceName, body: body);
  }

  ///Handle Job Result
  ///@param instance_name
  @POST(path: '/api/v1/job/result', optionalBody: true)
  Future<chopper.Response> _apiV1JobResultPost({
    @Query('instance_name') String? instanceName,
    @Body() required WorkerResult? body,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: '',
      summary: 'Handle Job Result',
      operationId: 'handle_job_result_api_v1_job_result_post',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Handlers"],
      deprecated: false,
    ),
  });

  ///Get all worker names
  Future<chopper.Response<List<String>>> apiV1WorkerGet() {
    return _apiV1WorkerGet();
  }

  ///Get all worker names
  @GET(path: '/api/v1/worker')
  Future<chopper.Response<List<String>>> _apiV1WorkerGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Retrieve a list of all available worker names.',
      summary: 'Get all worker names',
      operationId: 'get_workers_api_v1_worker_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Worker"],
      deprecated: false,
    ),
  });

  ///Get the configuration of a specific worker
  ///@param name
  Future<chopper.Response<NodeConfigExampleResponse>> apiV1WorkerNameGet({
    required String? name,
  }) {
    generatedMapping.putIfAbsent(
      NodeConfigExampleResponse,
      () => NodeConfigExampleResponse.fromJsonFactory,
    );

    return _apiV1WorkerNameGet(name: name);
  }

  ///Get the configuration of a specific worker
  ///@param name
  @GET(path: '/api/v1/worker/{name}')
  Future<chopper.Response<NodeConfigExampleResponse>> _apiV1WorkerNameGet({
    @Path('name') required String? name,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Retrieve the configuration of a specific worker by name.',
      summary: 'Get the configuration of a specific worker',
      operationId: 'get_worker_config_api_v1_worker__name__get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Worker"],
      deprecated: false,
    ),
  });

  ///Get all node definitions
  Future<chopper.Response<List<NodeDefinition>>> apiV1NodesGet() {
    generatedMapping.putIfAbsent(
      NodeDefinition,
      () => NodeDefinition.fromJsonFactory,
    );

    return _apiV1NodesGet();
  }

  ///Get all node definitions
  @GET(path: '/api/v1/nodes')
  Future<chopper.Response<List<NodeDefinition>>> _apiV1NodesGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description:
          'Returns metadata for all available nodes (triggers and workers). Includes port definitions, configuration fields, and UI properties.',
      summary: 'Get all node definitions',
      operationId: 'get_all_nodes_api_v1_nodes_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Nodes"],
      deprecated: false,
    ),
  });

  ///Get all trigger nodes
  Future<chopper.Response<List<NodeDefinition>>> apiV1NodesTriggersGet() {
    generatedMapping.putIfAbsent(
      NodeDefinition,
      () => NodeDefinition.fromJsonFactory,
    );

    return _apiV1NodesTriggersGet();
  }

  ///Get all trigger nodes
  @GET(path: '/api/v1/nodes/triggers')
  Future<chopper.Response<List<NodeDefinition>>> _apiV1NodesTriggersGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Returns all available trigger node definitions.',
      summary: 'Get all trigger nodes',
      operationId: 'get_trigger_nodes_api_v1_nodes_triggers_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Nodes"],
      deprecated: false,
    ),
  });

  ///Get all worker nodes
  Future<chopper.Response<List<NodeDefinition>>> apiV1NodesWorkersGet() {
    generatedMapping.putIfAbsent(
      NodeDefinition,
      () => NodeDefinition.fromJsonFactory,
    );

    return _apiV1NodesWorkersGet();
  }

  ///Get all worker nodes
  @GET(path: '/api/v1/nodes/workers')
  Future<chopper.Response<List<NodeDefinition>>> _apiV1NodesWorkersGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Returns all available worker node definitions.',
      summary: 'Get all worker nodes',
      operationId: 'get_worker_nodes_api_v1_nodes_workers_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Nodes"],
      deprecated: false,
    ),
  });

  ///Get nodes by category
  ///@param category
  Future<chopper.Response<List<NodeDefinition>>> apiV1NodesCategoryCategoryGet({
    required String? category,
  }) {
    generatedMapping.putIfAbsent(
      NodeDefinition,
      () => NodeDefinition.fromJsonFactory,
    );

    return _apiV1NodesCategoryCategoryGet(category: category);
  }

  ///Get nodes by category
  ///@param category
  @GET(path: '/api/v1/nodes/category/{category}')
  Future<chopper.Response<List<NodeDefinition>>>
  _apiV1NodesCategoryCategoryGet({
    @Path('category') required String? category,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description:
          'Returns all nodes in a specific category (trigger, scanner, analyzer, utility, output).',
      summary: 'Get nodes by category',
      operationId: 'get_nodes_by_category_api_v1_nodes_category__category__get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Nodes"],
      deprecated: false,
    ),
  });

  ///Get node definition by ID
  ///@param node_id
  Future<chopper.Response<NodeDefinition>> apiV1NodesNodeIdGet({
    required String? nodeId,
  }) {
    generatedMapping.putIfAbsent(
      NodeDefinition,
      () => NodeDefinition.fromJsonFactory,
    );

    return _apiV1NodesNodeIdGet(nodeId: nodeId);
  }

  ///Get node definition by ID
  ///@param node_id
  @GET(path: '/api/v1/nodes/{node_id}')
  Future<chopper.Response<NodeDefinition>> _apiV1NodesNodeIdGet({
    @Path('node_id') required String? nodeId,
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: 'Returns metadata for a specific node type.',
      summary: 'Get node definition by ID',
      operationId: 'get_node_api_v1_nodes__node_id__get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Nodes"],
      deprecated: false,
    ),
  });

  ///Health Check
  Future<chopper.Response> healthGet() {
    return _healthGet();
  }

  ///Health Check
  @GET(path: '/health')
  Future<chopper.Response> _healthGet({
    @chopper.Tag()
    SwaggerMetaData swaggerMetaData = const SwaggerMetaData(
      description: '',
      summary: 'Health Check',
      operationId: 'health_check_health_get',
      consumes: [],
      produces: [],
      security: [],
      tags: ["Health"],
      deprecated: false,
    ),
  });
}

typedef $JsonFactory<T> = T Function(Map<String, dynamic> json);

class $CustomJsonDecoder {
  $CustomJsonDecoder(this.factories);

  final Map<Type, $JsonFactory> factories;

  dynamic decode<T>(dynamic entity) {
    if (entity is Iterable) {
      return _decodeList<T>(entity);
    }

    if (entity is T) {
      return entity;
    }

    if (isTypeOf<T, Map>()) {
      return entity;
    }

    if (isTypeOf<T, Iterable>()) {
      return entity;
    }

    if (entity is Map<String, dynamic>) {
      return _decodeMap<T>(entity);
    }

    return entity;
  }

  T _decodeMap<T>(Map<String, dynamic> values) {
    final jsonFactory = factories[T];
    if (jsonFactory == null || jsonFactory is! $JsonFactory<T>) {
      return throw "Could not find factory for type $T. Is '$T: $T.fromJsonFactory' included in the CustomJsonDecoder instance creation in bootstrapper.dart?";
    }

    return jsonFactory(values);
  }

  List<T> _decodeList<T>(Iterable values) =>
      values.where((v) => v != null).map<T>((v) => decode<T>(v) as T).toList();
}

class $JsonSerializableConverter extends chopper.JsonConverter {
  @override
  FutureOr<chopper.Response<ResultType>> convertResponse<ResultType, Item>(
    chopper.Response response,
  ) async {
    if (response.bodyString.isEmpty) {
      // In rare cases, when let's say 204 (no content) is returned -
      // we cannot decode the missing json with the result type specified
      return chopper.Response(response.base, null, error: response.error);
    }

    if (ResultType == String) {
      return response.copyWith();
    }

    if (ResultType == DateTime) {
      return response.copyWith(
        body:
            DateTime.parse((response.body as String).replaceAll('"', ''))
                as ResultType,
      );
    }

    final jsonRes = await super.convertResponse(response);
    return jsonRes.copyWith<ResultType>(
      body: $jsonDecoder.decode<Item>(jsonRes.body) as ResultType,
    );
  }
}

final $jsonDecoder = $CustomJsonDecoder(generatedMapping);
