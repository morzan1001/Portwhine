// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format width=80

part of 'portwhine.swagger.dart';

// **************************************************************************
// ChopperGenerator
// **************************************************************************

// coverage:ignore-file
// ignore_for_file: type=lint
final class _$Portwhine extends Portwhine {
  _$Portwhine([ChopperClient? client]) {
    if (client == null) return;
    this.client = client;
  }

  @override
  final Type definitionType = Portwhine;

  @override
  Future<Response<List<String>>> _apiV1TriggerGet({
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
  }) {
    final Uri $url = Uri.parse('/api/v1/trigger');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<String>, String>($request);
  }

  @override
  Future<Response<NodeConfigExampleResponse>> _apiV1TriggerNameGet({
    required String? name,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/trigger/${name}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<NodeConfigExampleResponse, NodeConfigExampleResponse>(
      $request,
    );
  }

  @override
  Future<Response<PipelineResponse>> _apiV1PipelinePost({
    required Pipeline? body,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline');
    final $body = body;
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      body: $body,
      tag: swaggerMetaData,
    );
    return client.send<PipelineResponse, PipelineResponse>($request);
  }

  @override
  Future<Response<PipelineResponse>> _apiV1PipelinePatch({
    required PipelinePatch? body,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline');
    final $body = body;
    final Request $request = Request(
      'PATCH',
      $url,
      client.baseUrl,
      body: $body,
      tag: swaggerMetaData,
    );
    return client.send<PipelineResponse, PipelineResponse>($request);
  }

  @override
  Future<Response<PipelineResponse>> _apiV1PipelinePipelineIdGet({
    required String? pipelineId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline/${pipelineId}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<PipelineResponse, PipelineResponse>($request);
  }

  @override
  Future<Response<DeleteResponse>> _apiV1PipelinePipelineIdDelete({
    required String? pipelineId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline/${pipelineId}');
    final Request $request = Request(
      'DELETE',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<DeleteResponse, DeleteResponse>($request);
  }

  @override
  Future<Response<List<PipelineListItem>>> _apiV1PipelinesGet({
    int? size,
    int? page,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipelines');
    final Map<String, dynamic> $params = <String, dynamic>{
      'size': size,
      'page': page,
    };
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      parameters: $params,
      tag: swaggerMetaData,
    );
    return client.send<List<PipelineListItem>, PipelineListItem>($request);
  }

  @override
  Future<Response<MessageResponse>> _apiV1PipelineStartPipelineIdPost({
    required String? pipelineId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline/start/${pipelineId}');
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<MessageResponse, MessageResponse>($request);
  }

  @override
  Future<Response<MessageResponse>> _apiV1PipelineStopPipelineIdPost({
    required String? pipelineId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline/stop/${pipelineId}');
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<MessageResponse, MessageResponse>($request);
  }

  @override
  Future<Response<MessageResponse>> _apiV1PipelineCleanupPipelineIdPost({
    required String? pipelineId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/pipeline/cleanup/${pipelineId}');
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<MessageResponse, MessageResponse>($request);
  }

  @override
  Future<Response<dynamic>> _apiV1JobResultPost({
    String? instanceName,
    required WorkerResult? body,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/job/result');
    final Map<String, dynamic> $params = <String, dynamic>{
      'instance_name': instanceName,
    };
    final $body = body;
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      body: $body,
      parameters: $params,
      tag: swaggerMetaData,
    );
    return client.send<dynamic, dynamic>($request);
  }

  @override
  Future<Response<List<String>>> _apiV1WorkerGet({
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
  }) {
    final Uri $url = Uri.parse('/api/v1/worker');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<String>, String>($request);
  }

  @override
  Future<Response<NodeConfigExampleResponse>> _apiV1WorkerNameGet({
    required String? name,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/worker/${name}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<NodeConfigExampleResponse, NodeConfigExampleResponse>(
      $request,
    );
  }

  @override
  Future<Response<List<NodeDefinition>>> _apiV1NodesGet({
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
  }) {
    final Uri $url = Uri.parse('/api/v1/nodes');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<NodeDefinition>, NodeDefinition>($request);
  }

  @override
  Future<Response<List<NodeDefinition>>> _apiV1NodesTriggersGet({
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
  }) {
    final Uri $url = Uri.parse('/api/v1/nodes/triggers');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<NodeDefinition>, NodeDefinition>($request);
  }

  @override
  Future<Response<List<NodeDefinition>>> _apiV1NodesWorkersGet({
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
  }) {
    final Uri $url = Uri.parse('/api/v1/nodes/workers');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<NodeDefinition>, NodeDefinition>($request);
  }

  @override
  Future<Response<List<NodeDefinition>>> _apiV1NodesCategoryCategoryGet({
    required String? category,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/nodes/category/${category}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<List<NodeDefinition>, NodeDefinition>($request);
  }

  @override
  Future<Response<NodeDefinition>> _apiV1NodesNodeIdGet({
    required String? nodeId,
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
  }) {
    final Uri $url = Uri.parse('/api/v1/nodes/${nodeId}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<NodeDefinition, NodeDefinition>($request);
  }

  @override
  Future<Response<dynamic>> _healthGet({
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
  }) {
    final Uri $url = Uri.parse('/health');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      tag: swaggerMetaData,
    );
    return client.send<dynamic, dynamic>($request);
  }
}
