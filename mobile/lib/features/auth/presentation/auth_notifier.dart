import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_exceptions.dart';
import '../data/auth_models.dart';
import '../domain/auth_repository.dart';

enum AuthStatus { unauthenticated, loading, authenticated, error }

class AuthState {
  const AuthState({
    this.status = AuthStatus.unauthenticated,
    this.user,
    this.errorMessage,
  });

  final AuthStatus status;
  final UserProfile? user;
  final String? errorMessage;

  AuthState copyWith({
    AuthStatus? status,
    UserProfile? user,
    String? errorMessage,
  }) {
    return AuthState(
      status: status ?? this.status,
      user: user ?? this.user,
      errorMessage: errorMessage,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._repository) : super(const AuthState()) {
    _restoreSession();
  }

  final AuthRepository _repository;

  Future<void> _restoreSession() async {
    state = state.copyWith(status: AuthStatus.loading);
    try {
      final user = await _repository.restoreSession();
      if (user != null) {
        state = AuthState(status: AuthStatus.authenticated, user: user);
      } else {
        state = const AuthState();
      }
    } catch (_) {
      state = const AuthState();
    }
  }

  Future<void> login(String email, String password) async {
    state = state.copyWith(status: AuthStatus.loading, errorMessage: null);
    try {
      final response = await _repository.login(
        LoginRequest(email: email, password: password),
      );
      state = AuthState(status: AuthStatus.authenticated, user: response.user);
    } on Object catch (e) {
      // Catch everything (AppException, PlatformException, StateError, ...)
      // and surface a descriptive message so the user can see on-screen what
      // failed instead of the spinner hanging forever.
      state = AuthState(
        status: AuthStatus.error,
        errorMessage: describeError(e),
      );
    }
  }

  Future<void> register(
    String email,
    String password,
    String displayName,
  ) async {
    state = state.copyWith(status: AuthStatus.loading, errorMessage: null);
    try {
      final response = await _repository.register(
        RegisterRequest(
          email: email,
          password: password,
          displayName: displayName,
        ),
      );
      state = AuthState(status: AuthStatus.authenticated, user: response.user);
    } on Object catch (e) {
      state = AuthState(
        status: AuthStatus.error,
        errorMessage: describeError(e),
      );
    }
  }

  Future<void> logout() async {
    await _repository.logout();
    state = const AuthState();
  }
}

final authNotifierProvider = StateNotifierProvider<AuthNotifier, AuthState>((
  ref,
) {
  return AuthNotifier(ref.watch(authRepositoryProvider));
});
