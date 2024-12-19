import 'package:portwhine/global/constants.dart';
import 'package:portwhine/models/node_output_model.dart';
import 'package:portwhine/models/node_position.dart';

class NodeModel {
  String id;
  String name;
  Map<String, Type> inputs;
  Map<String, Type> outputs;
  Map<String, String> inputNodes;
  String? result;
  String? code;
  List<NodeOutputModel>? nodeOutputs;
  NodePosition? position;

  NodeModel({
    this.id = '',
    this.name = '',
    this.inputs = const {},
    this.outputs = const {},
    this.inputNodes = const {},
    this.result = nodeResult,
    this.code,
    this.nodeOutputs,
    this.position,
  });

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'name': name,
      'inputs': inputs,
      'outputs': outputs,
      'inputNodes': inputNodes,
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
      result: map['result'],
      code: map['code'],
      position: map['position'],
      nodeOutputs: map['nodeOutputs'] != null
          ? List<NodeOutputModel>.from(map['nodeOutputs']
              .map((output) => NodeOutputModel.fromMap(output)))
          : null,
    );
  }

  NodeModel copyWith({
    String? id,
    String? name,
    Map<String, Type>? inputs,
    Map<String, Type>? outputs,
    Map<String, String>? inputNodes,
    String? result,
    String? code,
    List<NodeOutputModel>? nodeOutputs,
    NodePosition? position,
  }) {
    return NodeModel(
      id: id ?? this.id,
      name: name ?? this.name,
      inputs: inputs ?? this.inputs,
      outputs: outputs ?? this.outputs,
      inputNodes: inputNodes ?? this.inputNodes,
      result: result ?? this.result,
      code: code ?? this.code,
      nodeOutputs: nodeOutputs ?? this.nodeOutputs,
      position: position ?? this.position,
    );
  }

  // Method to save node configuration
  Map<String, dynamic> saveConfiguration() {
    return toMap();
  }

  // Method to load node configuration
  static NodeModel loadConfiguration(Map<String, dynamic> map) {
    return fromMap(map);
  }

  // Method to validate connections between nodes
  bool validateConnections() {
    // Add logic to validate connections
    return true;
  }

  // Method to undo an action
  void undo() {
    // Add logic to undo an action
  }

  // Method to redo an action
  void redo() {
    // Add logic to redo an action
  }
}
