import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/api/api_client.dart';
import '../features/auth/presentation/login_page.dart';
import '../features/auth/presentation/onboarding_page.dart';
import '../features/auth/presentation/register_page.dart';
import '../features/profile/presentation/profile_page.dart';
import '../features/ranking/presentation/product_search_page.dart';
import '../features/receipt/presentation/receipt_confirm_page.dart';
import '../features/receipt/presentation/receipt_list_page.dart';
import '../features/receipt/presentation/receipt_review_page.dart';
import '../features/receipt/presentation/receipt_upload_page.dart';
import '../shared/widgets/app_scaffold.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final secureStorage = ref.watch(secureStorageProvider);

  return GoRouter(
    initialLocation: '/onboarding',
    redirect: (context, state) async {
      final path = state.matchedLocation;
      final isPublicPath =
          path == '/onboarding' || path == '/login' || path == '/register';

      if (isPublicPath) {
        final token = await secureStorage.readToken();
        if (token != null && token.isNotEmpty) {
          return '/home';
        }
        return null;
      }

      final token = await secureStorage.readToken();
      if (token == null || token.isEmpty) {
        return '/login';
      }
      return null;
    },
    routes: [
      GoRoute(
        path: '/onboarding',
        builder: (context, state) => const OnboardingPage(),
      ),
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterPage(),
      ),
      ShellRoute(
        builder: (context, state, child) => AppScaffold(child: child),
        routes: [
          GoRoute(
            path: '/home',
            builder: (context, state) => const ReceiptUploadPage(),
          ),
          GoRoute(
            path: '/receipts',
            builder: (context, state) => const ReceiptListPage(),
            routes: [
              GoRoute(
                path: ':id',
                builder: (context, state) =>
                    ReceiptReviewPage(receiptId: state.pathParameters['id']!),
                routes: [
                  GoRoute(
                    path: 'confirm',
                    builder: (context, state) => ReceiptConfirmPage(
                      receiptId: state.pathParameters['id']!,
                    ),
                  ),
                ],
              ),
            ],
          ),
          GoRoute(
            path: '/search',
            builder: (context, state) => const ProductSearchPage(),
          ),
          GoRoute(
            path: '/profile',
            builder: (context, state) => const ProfilePage(),
          ),
        ],
      ),
    ],
  );
});
