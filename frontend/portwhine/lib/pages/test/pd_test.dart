import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:portwhine/pages/test/visual_widget.dart';

@RoutePage()
class PDTestPage extends StatelessWidget {
  const PDTestPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const PipelineVisualization(
      pipelineData: {
        'id': '3483401e-a43b-46f2-b6a0-bc0a4a36b90e',
        'status': 'Stopped',
        'name': 'Testpipeline Certstream Test',
        'trigger': {
          'CertstreamTrigger': {
            'regex': '//.example//.tlr',
            'id': '779890d0-5bb7-4c89-aa59-59759ab1de6c',
            'status': 'Stopped',
            'output': ['http']
          }
        },
        'worker': [
          {
            'ResolverWorker': {
              'id': '6c0d448c-a9c3-4937-93cf-d32ed87a4ed6',
              'status': 'Stopped',
              'input': ['http'],
              'output': ['ip'],
              'children': [
                {
                  'NmapWorker': {
                    'id': 'b2c37fad-1eea-4c09-9096-641861e22741',
                    'status': 'Stopped',
                    'input': ['ip'],
                    'output': ['ip', 'http'],
                    'children': [
                      {
                        'WebAppAnalyzerWorker': {
                          'id': 'a304c9dc-2613-4942-9fab-f3c9aa4092c8',
                          'status': 'Stopped',
                          'input': ['http'],
                          'output': ['http'],
                          'children': null
                        }
                      }
                    ]
                  }
                },
                {
                  'TestSSLWorker': {
                    'id': 'b24e0c80-7bec-475c-9f9d-e1757285e9b0',
                    'status': 'Stopped',
                    'input': ['ip'],
                    'output': ['ip'],
                    'children': null
                  }
                }
              ],
              'use_internal': false
            }
          },
          {
            'ScreenshotWorker': {
              'id': 'a76bfa7b-9532-43d5-a84a-7376c9236909',
              'status': 'Stopped',
              'input': ['http'],
              'output': ['http'],
              'children': null
            }
          }
        ]
      },
    );
  }
}
