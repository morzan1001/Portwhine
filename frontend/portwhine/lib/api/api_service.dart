// import 'dart:io';

import 'package:chopper/chopper.dart';

part 'api_service.chopper.dart';

@ChopperApi()
abstract class ApiService extends ChopperService {
  // Pipelines Endpoints
  @Get(path: '/pipelines')
  Future<Response<List<Map<String, dynamic>>>> getAllPipelines({
    @Query('size') int size = 10,
    @Query('page') int page = 1,
  });

  @Post(path: '/pipeline')
  Future<Response<Map<String, dynamic>>> createPipeline(
    @Body() Map<String, dynamic> pipelineInput,
  );

  @Delete(path: '/pipeline/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> deletePipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @Post(path: '/pipeline/start/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> startPipeline(
    @Path('pipeline_id') String pipelineId,
  );

  @Post(path: '/pipeline/stop/{pipeline_id}')
  Future<Response<Map<String, dynamic>>> stopPipeline(
    @Path('pipeline_id') String pipelineId,
  );

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
