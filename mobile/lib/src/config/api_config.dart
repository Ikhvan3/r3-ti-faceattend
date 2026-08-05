class ApiConfig {
  const ApiConfig({required this.baseUrl});

  factory ApiConfig.fromEnvironment() {
    return const ApiConfig(
      baseUrl: String.fromEnvironment(
        'API_BASE_URL',
        defaultValue: 'http://10.0.2.2:8080/api/v1',
      ),
    );
  }

  final String baseUrl;

  String buildUrl(String path) {
    final normalizedBase = baseUrl.trim().replaceAll(RegExp(r'/+$'), '');
    final normalizedPath = path.trim().replaceAll(RegExp(r'^/+'), '');
    if (normalizedPath.isEmpty) {
      return normalizedBase;
    }
    return '$normalizedBase/$normalizedPath';
  }
}
