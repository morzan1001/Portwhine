import 'package:portwhine/models/node_model.dart';

class PipelineErrorParser {
  /// Parses the error string from the backend and returns a map of
  /// Node ID -> Error Message for nodes that have errors.
  static Map<String, String> parseErrors(
    String errorString,
    List<NodeModel> nodes,
  ) {
    final Map<String, String> nodeErrors = {};

    // Check for "Invalid <NodeType> configuration" pattern
    // Example: "Value error, Invalid FFUFWorker configuration: ..."
    final regex = RegExp(r'Invalid (\w+) configuration');
    final matches = regex.allMatches(errorString);

    for (final match in matches) {
      final nodeType = match.group(1);
      if (nodeType != null) {
        // Find nodes with this type (definition ID or name)
        final matchingNodes = nodes.where(
          (n) => n.definition?.id == nodeType || n.name == nodeType,
        );

        for (final node in matchingNodes) {
          // We assign a generic error message for now, as we can't easily
          // distinguish which specific instance failed if there are multiple.
          // We also try to extract a bit more context if possible.

          // Try to extract the specific error details following the configuration error
          // This is a bit heuristic.
          String detail = 'Configuration error';
          final errorPart = errorString.substring(match.end);
          final nextErrorIndex = errorPart.indexOf('Value error,');

          if (nextErrorIndex != -1) {
            detail = errorPart.substring(0, nextErrorIndex).trim();
          } else {
            detail = errorPart.trim();
          }

          // Truncate if too long
          if (detail.length > 100) {
            detail = '${detail.substring(0, 100)}...';
          }

          // Clean up common prefixes
          detail = detail.replaceAll(RegExp(r'^:\s*'), '');
          detail = detail.replaceAll(
            RegExp(r'^\d+ validation errors? for \w+\s*'),
            '',
          );

          nodeErrors[node.id] = detail.isNotEmpty
              ? detail
              : 'Invalid configuration';
        }
      }
    }

    return nodeErrors;
  }
}
