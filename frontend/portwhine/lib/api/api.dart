/// Main API module - exports all API functionality.
///
/// This file serves as the single entry point for all API-related imports.
/// Use `import 'package:portwhine/api/api.dart';` to access all API features.
library;

import 'package:chopper/chopper.dart';

import 'package:portwhine/api/generated/portwhine.swagger.dart';

// Export generated types and client
export 'generated/portwhine.swagger.dart';

// Export WebSocket service
export 'websocket_service.dart';

/// The base URL for the API.
const String apiBaseUrl = 'https://api.portwhine.local';

/// Global API client instance.
late Portwhine _apiClient;

/// Get the global API client instance.
Portwhine get api => _apiClient;

/// Initialize the API client.
///
/// Call this once at app startup (e.g., in main.dart).
Future<void> initApi() async {
  _apiClient = Portwhine.create(
    baseUrl: Uri.parse(apiBaseUrl),
    converter: $JsonSerializableConverter(),
  );
}

/// Create a new API client instance.
///
/// Use this when you need a separate client instance,
/// e.g., for parallel requests or different configurations.
Portwhine createApiClient({
  Uri? baseUrl,
  Authenticator? authenticator,
  List<Interceptor>? interceptors,
}) {
  return Portwhine.create(
    baseUrl: baseUrl ?? Uri.parse(apiBaseUrl),
    authenticator: authenticator,
    interceptors: interceptors,
    converter: $JsonSerializableConverter(),
  );
}
