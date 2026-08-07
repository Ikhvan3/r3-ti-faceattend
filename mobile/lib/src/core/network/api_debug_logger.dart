import 'package:flutter/foundation.dart';

void logApiRequest(String method, String path) {
  if (!kDebugMode) {
    return;
  }
  debugPrint('API request method=$method path=$path');
}

void logApiResponse(String method, String path, int statusCode) {
  if (!kDebugMode) {
    return;
  }
  debugPrint('API response method=$method path=$path status=$statusCode');
}

void logApiException(String method, String path, String category) {
  if (!kDebugMode) {
    return;
  }
  debugPrint('API exception method=$method path=$path category=$category');
}
