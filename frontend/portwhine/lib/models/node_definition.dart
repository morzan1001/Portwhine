// Node definition models for rendering nodes dynamically from backend metadata.
//
// These models mirror the backend's NodeDefinition structure and provide
// all information needed to render and configure nodes in the UI.

import 'package:flutter/material.dart';
import 'package:portwhine/api/api.dart' as gen;

/// Supported field types for node configuration.
enum FieldType {
  string,
  integer,
  float,
  boolean,
  select,
  multiselect,
  ipAddress,
  ipList,
  regex,
  portRange;

  static FieldType fromString(String value) {
    switch (value) {
      case 'string':
        return FieldType.string;
      case 'integer':
        return FieldType.integer;
      case 'float':
        return FieldType.float;
      case 'boolean':
        return FieldType.boolean;
      case 'select':
        return FieldType.select;
      case 'multiselect':
        return FieldType.multiselect;
      case 'ip_address':
        return FieldType.ipAddress;
      case 'ip_list':
        return FieldType.ipList;
      case 'regex':
        return FieldType.regex;
      case 'port_range':
        return FieldType.portRange;
      default:
        return FieldType.string;
    }
  }

  static FieldType fromGenerated(gen.FieldType type) {
    switch (type) {
      case gen.FieldType.string:
        return FieldType.string;
      case gen.FieldType.integer:
        return FieldType.integer;
      case gen.FieldType.float:
        return FieldType.float;
      case gen.FieldType.boolean:
        return FieldType.boolean;
      case gen.FieldType.select:
        return FieldType.select;
      case gen.FieldType.multiselect:
        return FieldType.multiselect;
      case gen.FieldType.ipAddress:
        return FieldType.ipAddress;
      case gen.FieldType.ipList:
        return FieldType.ipList;
      case gen.FieldType.regex:
        return FieldType.regex;
      case gen.FieldType.portRange:
        return FieldType.portRange;
      case gen.FieldType.swaggerGeneratedUnknown:
        return FieldType.string;
    }
  }
}

/// Definition of a configuration field for a node.
class FieldDefinition {
  final String name;
  final String label;
  final FieldType type;
  final String description;
  final bool required;
  final dynamic defaultValue;
  final List<String>? options;
  final String? placeholder;
  final String? validationPattern;

  const FieldDefinition({
    required this.name,
    required this.label,
    required this.type,
    this.description = '',
    this.required = false,
    this.defaultValue,
    this.options,
    this.placeholder,
    this.validationPattern,
  });

  factory FieldDefinition.fromJson(Map<String, dynamic> json) {
    return FieldDefinition(
      name: json['name'] as String,
      label: json['label'] as String,
      type: FieldType.fromString(json['type'] as String),
      description: json['description'] as String? ?? '',
      required: json['required'] as bool? ?? false,
      defaultValue: json['default'],
      options: (json['options'] as List?)?.cast<String>(),
      placeholder: json['placeholder'] as String?,
      validationPattern: json['validation_pattern'] as String?,
    );
  }

  factory FieldDefinition.fromGenerated(gen.FieldDefinition fd) {
    return FieldDefinition(
      name: fd.name,
      label: fd.label,
      type: FieldType.fromGenerated(fd.type),
      description: fd.description ?? '',
      required: fd.required ?? false,
      defaultValue: fd.$default,
      options: fd.options,
      placeholder: fd.placeholder,
      validationPattern: fd.validationPattern,
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'label': label,
        'type': type.name,
        'description': description,
        'required': required,
        'default': defaultValue,
        'options': options,
        'placeholder': placeholder,
        'validation_pattern': validationPattern,
      };
}

/// Type of data that flows through ports.
enum DataType {
  http,
  ip;

  static DataType fromString(String value) {
    switch (value.toLowerCase()) {
      case 'http':
        return DataType.http;
      case 'ip':
        return DataType.ip;
      default:
        return DataType.ip;
    }
  }

  static DataType fromGenerated(gen.InputOutputType type) {
    switch (type) {
      case gen.InputOutputType.http:
        return DataType.http;
      case gen.InputOutputType.ip:
        return DataType.ip;
      case gen.InputOutputType.swaggerGeneratedUnknown:
        return DataType.ip;
    }
  }

  String get displayName {
    switch (this) {
      case DataType.http:
        return 'HTTP';
      case DataType.ip:
        return 'IP';
    }
  }

  Color get color {
    switch (this) {
      case DataType.http:
        return const Color(0xFFF59E0B); // Amber
      case DataType.ip:
        return const Color(0xFF10B981); // Emerald
    }
  }
}

/// Definition of an input or output port on a node.
class PortDefinition {
  final String id;
  final String label;
  final DataType dataType;
  final String description;
  final bool required;
  final bool multiple;

  const PortDefinition({
    required this.id,
    required this.label,
    required this.dataType,
    this.description = '',
    this.required = true,
    this.multiple = false,
  });

  factory PortDefinition.fromJson(Map<String, dynamic> json) {
    return PortDefinition(
      id: json['id'] as String,
      label: json['label'] as String,
      dataType: DataType.fromString(json['data_type'] as String),
      description: json['description'] as String? ?? '',
      required: json['required'] as bool? ?? true,
      multiple: json['multiple'] as bool? ?? false,
    );
  }

  factory PortDefinition.fromGenerated(gen.PortDefinition pd) {
    return PortDefinition(
      id: pd.id,
      label: pd.label,
      dataType: DataType.fromGenerated(pd.dataType),
      description: pd.description ?? '',
      required: pd.required ?? true,
      multiple: pd.multiple ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'label': label,
        'data_type': dataType.name,
        'description': description,
        'required': required,
        'multiple': multiple,
      };
}

/// Type of node - trigger or worker
enum NodeType {
  trigger,
  worker;

  static NodeType fromString(String value) {
    switch (value.toLowerCase()) {
      case 'trigger':
        return NodeType.trigger;
      case 'worker':
        return NodeType.worker;
      default:
        return NodeType.worker;
    }
  }

  static NodeType fromGenerated(gen.NodeType type) {
    switch (type) {
      case gen.NodeType.trigger:
        return NodeType.trigger;
      case gen.NodeType.worker:
        return NodeType.worker;
      case gen.NodeType.swaggerGeneratedUnknown:
        return NodeType.worker;
    }
  }

  bool get isTrigger => this == NodeType.trigger;
  bool get isWorker => this == NodeType.worker;
}

/// Categories for organizing worker nodes in the UI.
/// Only applies to workers - triggers don't have categories.
enum WorkerCategory {
  scanner,
  analyzer,
  utility,
  output;

  static WorkerCategory? fromString(String? value) {
    if (value == null) return null;
    switch (value.toLowerCase()) {
      case 'scanner':
        return WorkerCategory.scanner;
      case 'analyzer':
        return WorkerCategory.analyzer;
      case 'utility':
        return WorkerCategory.utility;
      case 'output':
        return WorkerCategory.output;
      default:
        return WorkerCategory.utility;
    }
  }

  static WorkerCategory? fromGenerated(gen.WorkerCategory? category) {
    if (category == null) return null;
    switch (category) {
      case gen.WorkerCategory.scanner:
        return WorkerCategory.scanner;
      case gen.WorkerCategory.analyzer:
        return WorkerCategory.analyzer;
      case gen.WorkerCategory.utility:
        return WorkerCategory.utility;
      case gen.WorkerCategory.output:
        return WorkerCategory.output;
      case gen.WorkerCategory.swaggerGeneratedUnknown:
        return WorkerCategory.utility;
    }
  }

  String get displayName {
    switch (this) {
      case WorkerCategory.scanner:
        return 'Scanners';
      case WorkerCategory.analyzer:
        return 'Analyzers';
      case WorkerCategory.utility:
        return 'Utilities';
      case WorkerCategory.output:
        return 'Outputs';
    }
  }

  IconData get icon {
    switch (this) {
      case WorkerCategory.scanner:
        return Icons.radar;
      case WorkerCategory.analyzer:
        return Icons.analytics_outlined;
      case WorkerCategory.utility:
        return Icons.build_outlined;
      case WorkerCategory.output:
        return Icons.output;
    }
  }
}

/// Complete definition of a node type from the backend.
class NodeDefinition {
  final String id;
  final String name;
  final String description;
  final NodeType nodeType;
  final WorkerCategory? category; // Only for workers, null for triggers
  final String icon;
  final Color color;
  final List<PortDefinition> inputs;
  final List<PortDefinition> outputs;
  final List<FieldDefinition> configFields;
  final String imageName;
  final bool supportsMultipleInstances;

  const NodeDefinition({
    required this.id,
    required this.name,
    required this.description,
    required this.nodeType,
    this.category,
    this.icon = 'default',
    this.color = const Color(0xFF6366F1),
    this.inputs = const [],
    this.outputs = const [],
    this.configFields = const [],
    required this.imageName,
    this.supportsMultipleInstances = true,
  });

  /// Whether this is a trigger node
  bool get isTrigger => nodeType == NodeType.trigger;

  /// Whether this is a worker node
  bool get isWorker => nodeType == NodeType.worker;

  factory NodeDefinition.fromJson(Map<String, dynamic> json) {
    return NodeDefinition(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      nodeType: NodeType.fromString(json['node_type'] as String? ?? 'worker'),
      category: WorkerCategory.fromString(json['category'] as String?),
      icon: json['icon'] as String? ?? 'default',
      color: _parseColor(json['color'] as String? ?? '#6366F1'),
      inputs: (json['inputs'] as List?)
              ?.map((e) => PortDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      outputs: (json['outputs'] as List?)
              ?.map((e) => PortDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      configFields: (json['config_fields'] as List?)
              ?.map((e) => FieldDefinition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      imageName: json['image_name'] as String,
      supportsMultipleInstances:
          json['supports_multiple_instances'] as bool? ?? true,
    );
  }

  /// Create from generated API type.
  factory NodeDefinition.fromGenerated(gen.NodeDefinition nd) {
    return NodeDefinition(
      id: nd.id,
      name: nd.name,
      description: nd.description,
      nodeType: NodeType.fromGenerated(nd.nodeType),
      category: WorkerCategory.fromGenerated(nd.category),
      icon: nd.icon ?? 'default',
      color: _parseColor(nd.color ?? '#6366F1'),
      inputs: nd.inputs?.map(PortDefinition.fromGenerated).toList() ?? [],
      outputs: nd.outputs?.map(PortDefinition.fromGenerated).toList() ?? [],
      configFields:
          nd.configFields?.map(FieldDefinition.fromGenerated).toList() ?? [],
      imageName: nd.imageName,
      supportsMultipleInstances: nd.supportsMultipleInstances ?? true,
    );
  }

  static Color _parseColor(String hexColor) {
    hexColor = hexColor.replaceFirst('#', '');
    if (hexColor.length == 6) {
      hexColor = 'FF$hexColor';
    }
    return Color(int.parse(hexColor, radix: 16));
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'description': description,
        'node_type': nodeType.name,
        'category': category?.name,
        'icon': icon,
        'color': '#${color.toARGB32().toRadixString(16).substring(2)}',
        'inputs': inputs.map((e) => e.toJson()).toList(),
        'outputs': outputs.map((e) => e.toJson()).toList(),
        'config_fields': configFields.map((e) => e.toJson()).toList(),
        'image_name': imageName,
        'supports_multiple_instances': supportsMultipleInstances,
      };

  /// Get the icon widget for this node type.
  IconData get iconData {
    switch (icon) {
      case 'network':
        return Icons.hub_outlined;
      case 'certificate':
        return Icons.verified_outlined;
      case 'radar':
        return Icons.radar;
      case 'dns':
        return Icons.dns_outlined;
      case 'search':
        return Icons.search;
      case 'shield':
        return Icons.shield_outlined;
      case 'camera':
        return Icons.camera_alt_outlined;
      case 'lock':
        return Icons.lock_outline;
      case 'code':
        return Icons.code;
      default:
        return Icons.extension;
    }
  }
}
