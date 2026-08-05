class UserProfile {
  const UserProfile({
    required this.id,
    required this.employeeNumber,
    required this.name,
    required this.email,
    required this.phone,
    required this.position,
    required this.role,
    required this.accountStatus,
  });

  factory UserProfile.fromJson(Object? value) {
    if (value is! Map<String, Object?>) {
      throw const FormatException('Profil user tidak valid.');
    }

    return UserProfile(
      id: _requiredString(value, 'id'),
      employeeNumber: _requiredString(value, 'employee_number'),
      name: _requiredString(value, 'name'),
      email: _requiredString(value, 'email'),
      phone: _nullableString(value, 'phone'),
      position: _nullableString(value, 'position'),
      role: _requiredString(value, 'role'),
      accountStatus: _requiredString(value, 'account_status'),
    );
  }

  final String id;
  final String employeeNumber;
  final String name;
  final String email;
  final String? phone;
  final String? position;
  final String role;
  final String accountStatus;

  bool get isUser => role == 'USER';
  bool get isActive => accountStatus == 'ACTIVE';
}

String _requiredString(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value is String && value.trim().isNotEmpty) {
    return value;
  }
  throw FormatException('Field $key tidak valid.');
}

String? _nullableString(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value == null) {
    return null;
  }
  if (value is String) {
    return value;
  }
  throw FormatException('Field $key tidak valid.');
}
