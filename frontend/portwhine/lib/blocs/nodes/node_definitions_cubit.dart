import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';

import 'package:portwhine/api/api.dart' as gen;
import 'package:portwhine/models/node_definition.dart';

part 'node_definitions_state.dart';

/// Cubit for managing node definitions loaded from the backend.
///
/// This cubit loads all available node definitions (triggers and workers)
/// from the API and caches them for use throughout the application.
class NodeDefinitionsCubit extends Cubit<NodeDefinitionsState> {
  NodeDefinitionsCubit() : super(NodeDefinitionsInitial());

  /// Load all node definitions from the API.
  Future<void> loadNodeDefinitions() async {
    emit(NodeDefinitionsLoading());

    try {
      final response = await gen.api.apiV1NodesGet();

      if (response.isSuccessful && response.body != null) {
        final definitions = response.body!
            .map((nd) => NodeDefinition.fromGenerated(nd))
            .toList();

        emit(NodeDefinitionsLoaded(definitions: definitions));
      } else {
        emit(NodeDefinitionsError(
          message: 'Failed to load node definitions: ${response.statusCode}',
        ));
      }
    } catch (e) {
      emit(NodeDefinitionsError(message: e.toString()));
    }
  }

  /// Get a node definition by its ID.
  NodeDefinition? getNodeById(String id) {
    final currentState = state;
    if (currentState is NodeDefinitionsLoaded) {
      try {
        return currentState.definitions.firstWhere((d) => d.id == id);
      } catch (_) {
        return null;
      }
    }
    return null;
  }

  /// Get all trigger node definitions.
  List<NodeDefinition> get triggers {
    final currentState = state;
    if (currentState is NodeDefinitionsLoaded) {
      return currentState.definitions.where((d) => d.isTrigger).toList();
    }
    return [];
  }

  /// Get all worker node definitions.
  List<NodeDefinition> get workers {
    final currentState = state;
    if (currentState is NodeDefinitionsLoaded) {
      return currentState.definitions.where((d) => d.isWorker).toList();
    }
    return [];
  }

  /// Get workers grouped by category.
  Map<WorkerCategory, List<NodeDefinition>> get workersByCategory {
    final currentState = state;
    if (currentState is NodeDefinitionsLoaded) {
      final map = <WorkerCategory, List<NodeDefinition>>{};
      for (final definition in currentState.definitions) {
        if (definition.category != null) {
          map.putIfAbsent(definition.category!, () => []).add(definition);
        }
      }
      return map;
    }
    return {};
  }
}
