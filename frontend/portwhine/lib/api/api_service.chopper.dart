// dart format width=80
// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'api_service.dart';

// **************************************************************************
// ChopperGenerator
// **************************************************************************

// coverage:ignore-file
// ignore_for_file: type=lint
final class _$ApiService extends ApiService {
  _$ApiService([ChopperClient? client]) {
    if (client == null) return;
    this.client = client;
  }

  @override
  final Type definitionType = ApiService;

  @override
  Future<Response<List<Map<String, dynamic>>>> getAllPipelines({
    int size = 10,
    int page = 1,
  }) {
    final Uri $url = Uri.parse('/pipelines');
    final Map<String, dynamic> $params = <String, dynamic>{
      'size': size,
      'page': page,
    };
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
      parameters: $params,
    );
    return client
        .send<List<Map<String, dynamic>>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> createPipeline(
      Map<String, dynamic> pipelineInput) {
    final Uri $url = Uri.parse('/pipeline');
    final $body = pipelineInput;
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
      body: $body,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> deletePipeline(String pipelineId) {
    final Uri $url = Uri.parse('/pipeline/${pipelineId}');
    final Request $request = Request(
      'DELETE',
      $url,
      client.baseUrl,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> startPipeline(String pipelineId) {
    final Uri $url = Uri.parse('/pipeline/start/${pipelineId}');
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> stopPipeline(String pipelineId) {
    final Uri $url = Uri.parse('/pipeline/stop/${pipelineId}');
    final Request $request = Request(
      'POST',
      $url,
      client.baseUrl,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> getPipeline(String pipelineId) {
    final Uri $url = Uri.parse('/pipeline/${pipelineId}');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<Map<String, dynamic>>> updatePipeline(
    String pipelineId,
    Map<String, dynamic> pipelineInput,
  ) {
    final Uri $url = Uri.parse('/pipeline/${pipelineId}');
    final $body = pipelineInput;
    final Request $request = Request(
      'PUT',
      $url,
      client.baseUrl,
      body: $body,
    );
    return client.send<Map<String, dynamic>, Map<String, dynamic>>($request);
  }

  @override
  Future<Response<List<String>>> getAllWorkers() {
    final Uri $url = Uri.parse('/worker');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
    );
    return client.send<List<String>, String>($request);
  }

  @override
  Future<Response<List<String>>> getAllTriggers() {
    final Uri $url = Uri.parse('/trigger');
    final Request $request = Request(
      'GET',
      $url,
      client.baseUrl,
    );
    return client.send<List<String>, String>($request);
  }
}
