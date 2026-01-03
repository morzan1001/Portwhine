part of 'node_definitions_cubit.dart';

/// Base state for node definitions.
sealed class NodeDefinitionsState extends Equatable {
  const NodeDefinitionsState();

  @override
  List<Object?> get props => [];
}

/// Initial state before loading.
final class NodeDefinitionsInitial extends NodeDefinitionsState {}

/// Loading state while fetching definitions.
final class NodeDefinitionsLoading extends NodeDefinitionsState {}

/// Loaded state with all definitions.
final class NodeDefinitionsLoaded extends NodeDefinitionsState {
  final List<NodeDefinition> definitions;

  const NodeDefinitionsLoaded({required this.definitions});

  @override
  List<Object?> get props => [definitions];

  /// Get definition by ID.
  NodeDefinition? getById(String id) {
    try {
      return definitions.firstWhere((d) => d.id == id);
    } catch (_) {
      return null;
    }
  }

  /// Get all triggers.
  List<NodeDefinition> get triggers =>
      definitions.where((d) => d.isTrigger).toList();

  /// Get all workers.
  List<NodeDefinition> get workers =>
      definitions.where((d) => !d.isTrigger).toList();
}

/// Error state if loading failed.
final class NodeDefinitionsError extends NodeDefinitionsState {
  final String message;

  const NodeDefinitionsError({required this.message});

  @override
  List<Object?> get props => [message];
}
