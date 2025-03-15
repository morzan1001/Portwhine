// import 'dart:io';

import 'package:chopper/chopper.dart';

part 'api_service.chopper.dart';

@ChopperApi()
abstract class ApiService extends ChopperService {
  // Pipelines Endpoints
  @GET(path: '/pipelines')
  Future<Response<List<Map<String, dynamic>>>> getAllPipelines({
    @Query('size') int size = 10,
    @Query('page') int page = 1,
  });

  @POST(path: '/pipeline')
  Future<Response<Map<String, dynamic>>> createPipeline(
    @Body() Map<String, dynamic> pipelineInput,
  );

  @DELETE(path: '/pipeline/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> deletePipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @POST(path: '/pipeline/start/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> startPipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @POST(path: '/pipeline/stop/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> stopPipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @GET(path: '/pipeline/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> getPipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @PUT(path: '/pipeline/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> updatePipeline(
    @Path('pipeline_id') String pipelineId,
    @Body() Map<String, dynamic> pipelineInput,
  );

  @GET(path: '/worker')
  Future<Response<List<String>>> getAllWorkers();

  @GET(path: '/trigger')
  Future<Response<List<String>>> getAllTriggers();

  // chopper client
  static Future<ApiService> create() async {
    final client = ChopperClient(
      baseUrl: Uri.parse('https://37.27.179.252:8000/api/v1/'),
      services: [_$ApiService()],
      interceptors: [],
      converter: const JsonConverter(),
    );

    return _$ApiService(client);
  }
}
