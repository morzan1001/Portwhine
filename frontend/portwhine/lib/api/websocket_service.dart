import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import 'package:portwhine/models/node_status.dart';

/// Service for managing WebSocket connections to receive real-time pipeline status updates.
class PipelineWebSocketService {
  static const String _baseUrl = 'wss://api.portwhine.local/api/v1';

  WebSocketChannel? _channel;
  StreamController<PipelineStatusUpdate>? _statusController;
  Timer? _pingTimer;
  Timer? _reconnectTimer;
  String? _currentPipelineId;
  bool _isConnected = false;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 5;
  static const Duration _reconnectDelay = Duration(seconds: 3);
  static const Duration _pingInterval = Duration(seconds: 30);

  /// Stream of status updates for the current pipeline.
  Stream<PipelineStatusUpdate>? get statusStream => _statusController?.stream;

  /// Whether the WebSocket is currently connected.
  bool get isConnected => _isConnected;

  /// Connect to a specific pipeline's status stream.
  Future<void> connectToPipeline(String pipelineId) async {
    // If already connected to the same pipeline, do nothing
    if (_currentPipelineId == pipelineId && _isConnected) {
      return;
    }

    // Disconnect from any existing connection
    await disconnect();

    _currentPipelineId = pipelineId;
    _statusController = StreamController<PipelineStatusUpdate>.broadcast();

    await _connect();
  }

  Future<void> _connect() async {
    if (_currentPipelineId == null) return;

    try {
      final uri = Uri.parse('$_baseUrl/ws/pipeline/$_currentPipelineId');
      _channel = WebSocketChannel.connect(uri);

      await _channel!.ready;
      _isConnected = true;
      _reconnectAttempts = 0;

      debugPrint('WebSocket connected to pipeline: $_currentPipelineId');

      // Listen for messages
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDisconnect,
        cancelOnError: false,
      );

      // Start ping timer to keep connection alive
      _startPingTimer();
    } catch (e) {
      debugPrint('WebSocket connection error: $e');
      _isConnected = false;
      _scheduleReconnect();
    }
  }

  void _handleMessage(dynamic message) {
    try {
      final data = jsonDecode(message as String) as Map<String, dynamic>;
      final wsMessage = WebSocketMessage.fromJson(data);

      switch (wsMessage.type) {
        case WebSocketMessageType.statusUpdate:
          final statusUpdate = wsMessage.asStatusUpdate;
          if (statusUpdate != null) {
            _statusController?.add(statusUpdate);
          }
          break;
        case WebSocketMessageType.pong:
          // Connection is alive, nothing to do
          break;
        case WebSocketMessageType.error:
          debugPrint('WebSocket error message: ${wsMessage.data}');
          break;
        default:
          break;
      }
    } catch (e) {
      debugPrint('Error parsing WebSocket message: $e');
    }
  }

  void _handleError(dynamic error) {
    debugPrint('WebSocket error: $error');
    _isConnected = false;
    _scheduleReconnect();
  }

  void _handleDisconnect() {
    debugPrint('WebSocket disconnected');
    _isConnected = false;
    _stopPingTimer();
    _scheduleReconnect();
  }

  void _startPingTimer() {
    _stopPingTimer();
    _pingTimer = Timer.periodic(_pingInterval, (_) {
      _sendPing();
    });
  }

  void _stopPingTimer() {
    _pingTimer?.cancel();
    _pingTimer = null;
  }

  void _sendPing() {
    if (_isConnected && _channel != null) {
      try {
        _channel!.sink.add(jsonEncode({'type': 'ping'}));
      } catch (e) {
        debugPrint('Error sending ping: $e');
      }
    }
  }

  void _scheduleReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      debugPrint('Max reconnect attempts reached');
      return;
    }

    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(_reconnectDelay, () {
      _reconnectAttempts++;
      debugPrint('Attempting to reconnect (attempt $_reconnectAttempts)');
      _connect();
    });
  }

  /// Subscribe to status updates for an additional pipeline.
  void subscribeToPipeline(String pipelineId) {
    if (_isConnected && _channel != null) {
      _channel!.sink.add(jsonEncode({
        'type': 'subscribe',
        'pipeline_id': pipelineId,
      }));
    }
  }

  /// Unsubscribe from a pipeline's status updates.
  void unsubscribeFromPipeline(String pipelineId) {
    if (_isConnected && _channel != null) {
      _channel!.sink.add(jsonEncode({
        'type': 'unsubscribe',
        'pipeline_id': pipelineId,
      }));
    }
  }

  /// Disconnect from the WebSocket.
  Future<void> disconnect() async {
    _stopPingTimer();
    _reconnectTimer?.cancel();
    _reconnectTimer = null;

    await _channel?.sink.close();
    _channel = null;

    await _statusController?.close();
    _statusController = null;

    _currentPipelineId = null;
    _isConnected = false;
    _reconnectAttempts = 0;

    debugPrint('WebSocket disconnected');
  }
}

/// Singleton instance of the WebSocket service.
final pipelineWebSocketService = PipelineWebSocketService();
