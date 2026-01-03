import 'package:portwhine/models/node_definition.dart';
import 'package:portwhine/models/node_output_model.dart';
import 'package:portwhine/models/node_position.dart';
import 'package:portwhine/models/node_status.dart';

class NodeModel {
  String id;
  String name;
  Map<String, Type> inputs;
  Map<String, Type> outputs;
  Map<String, String> inputNodes;
  Map<String, dynamic> config;
  String? result;
  String? code;
  List<NodeOutputModel>? nodeOutputs;
  NodePosition? position;

  /// Definition loaded from backend (contains ports, description, color, etc.)
  NodeDefinition? definition;

  /// Current runtime status of the node
  NodeStatusInfo? statusInfo;

  /// Error message associated with this node (e.g. validation error)
  String error;

  NodeModel({
    this.id = '',
    this.name = '',
    this.inputs = const {},
    this.outputs = const {},
    this.inputNodes = const {},
    this.config = const {},
    this.result = '',
    this.code,
    this.nodeOutputs,
    this.position,
    this.definition,
    this.statusInfo,
    this.error = '',
  });

  /// Whether this node is a trigger (has no inputs).
  bool get isTrigger => definition?.isTrigger ?? inputs.isEmpty;

  /// Get the list of input ports from the definition.
  List<PortDefinition> get inputPorts => definition?.inputs ?? [];

  /// Get the list of output ports from the definition.
  List<PortDefinition> get outputPorts => definition?.outputs ?? [];

  /// Get the node's description from the definition.
  String get description => definition?.description ?? '';

  /// Get the current status.
  NodeStatus get status => statusInfo?.status ?? NodeStatus.unknown;

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'name': name,
      'inputs': inputs,
      'outputs': outputs,
      'inputNodes': inputNodes,
      'config': config,
      'result': result,
      'code': code,
      'nodeOutputs': nodeOutputs?.map((output) => output.toMap()).toList(),
      'position': position,
    };
  }

  static NodeModel fromMap(Map<String, dynamic> map) {
    return NodeModel(
      id: map['id'],
      name: map['name'],
      inputs: map['inputs'],
      outputs: map['outputs'],
      inputNodes: map['inputNodes'],
      config: map['config'] ?? {},
      result: map['result'],
      code: map['code'],
      position: map['position'],
      nodeOutputs: map['nodeOutputs'] != null
          ? List<NodeOutputModel>.from(
              map['nodeOutputs'].map(
                (output) => NodeOutputModel.fromMap(output),
              ),
            )
          : null,
    );
  }

  /// Create a NodeModel from a NodeDefinition (used for list display).
  factory NodeModel.fromNodeDefinition(NodeDefinition definition) {
    return NodeModel(
      id: '', // Empty ID so it gets generated when added to canvas
      name: definition.id,
      inputs: {for (var p in definition.inputs) p.id: String},
      outputs: {for (var p in definition.outputs) p.id: String},
      definition: definition,
    );
  }

  NodeModel copyWith({
    String? id,
    String? name,
    Map<String, Type>? inputs,
    Map<String, Type>? outputs,
    Map<String, String>? inputNodes,
    Map<String, dynamic>? config,
    String? result,
    String? code,
    List<NodeOutputModel>? nodeOutputs,
    NodePosition? position,
    NodeDefinition? definition,
    NodeStatusInfo? statusInfo,
    String? error,
  }) {
    return NodeModel(
      id: id ?? this.id,
      name: name ?? this.name,
      inputs: inputs ?? this.inputs,
      outputs: outputs ?? this.outputs,
      inputNodes: inputNodes ?? this.inputNodes,
      config: config ?? this.config,
      result: result ?? this.result,
      code: code ?? this.code,
      nodeOutputs: nodeOutputs ?? this.nodeOutputs,
      position: position ?? this.position,
      definition: definition ?? this.definition,
      statusInfo: statusInfo ?? this.statusInfo,
      error: error ?? this.error,
    );
  }
}
