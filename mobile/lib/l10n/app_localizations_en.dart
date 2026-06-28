// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appName => 'AhorraApp';

  @override
  String get onboardingTitle1 => 'Find where to buy cheaper';

  @override
  String get onboardingSubtitle1 =>
      'Scan your receipts and discover the best prices near you.';

  @override
  String get onboardingTitle2 => 'Earn points';

  @override
  String get onboardingSubtitle2 =>
      'Get rewarded every time you confirm a receipt.';

  @override
  String get onboardingTitle3 => 'Join the community';

  @override
  String get onboardingSubtitle3 =>
      'Help others save money by sharing price data.';

  @override
  String get skip => 'Skip';

  @override
  String get next => 'Next';

  @override
  String get getStarted => 'Get Started';

  @override
  String get login => 'Login';

  @override
  String get register => 'Register';

  @override
  String get email => 'Email';

  @override
  String get password => 'Password';

  @override
  String get displayName => 'Display Name';

  @override
  String get loginSubtitle => 'Welcome back! Please log in to continue.';

  @override
  String get registerSubtitle => 'Create an account to start saving.';

  @override
  String get noAccount => 'Don\'t have an account?';

  @override
  String get haveAccount => 'Already have an account?';

  @override
  String get scanReceipt => 'Scan Receipt';

  @override
  String get uploading => 'Uploading...';

  @override
  String retrying(Object attempt, Object max) {
    return 'Retrying ($attempt/$max)...';
  }

  @override
  String get retry => 'Retry';

  @override
  String get receiptPending => 'Processing receipt...';

  @override
  String get receiptNeedsReview => 'Needs Review';

  @override
  String get receiptConfirmed => 'Confirmed';

  @override
  String get receiptRejected => 'Rejected';

  @override
  String get reviewReceipt => 'Review Receipt';

  @override
  String get store => 'Store';

  @override
  String get storeName => 'Store Name';

  @override
  String get branch => 'Branch';

  @override
  String get address => 'Address';

  @override
  String get purchaseDate => 'Purchase Date';

  @override
  String get total => 'Total';

  @override
  String get items => 'Items';

  @override
  String get addItem => 'Add Item';

  @override
  String get product => 'Product';

  @override
  String get quantity => 'Quantity';

  @override
  String get price => 'Price';

  @override
  String get currency => 'Currency';

  @override
  String get confirm => 'Confirm';

  @override
  String get reject => 'Reject';

  @override
  String get pointsEarned => 'Points Earned';

  @override
  String get searchProducts => 'Search Products';

  @override
  String get searchHint => 'Search for a product...';

  @override
  String get noResults => 'No results found';

  @override
  String noResultsFor(Object query) {
    return 'No results found for \'$query\'';
  }

  @override
  String get cheapestStore => 'Cheapest Store';

  @override
  String get profile => 'Profile';

  @override
  String get totalPoints => 'Total Points';

  @override
  String get recentActivity => 'Recent Activity';

  @override
  String get logout => 'Logout';

  @override
  String get home => 'Home';

  @override
  String get receipts => 'Receipts';

  @override
  String get currencyRequired => 'Currency is required for all items';

  @override
  String get emailRequired => 'Email is required';

  @override
  String get passwordRequired => 'Password is required';

  @override
  String get passwordMinLength => 'Password must be at least 8 characters';

  @override
  String get nameRequired => 'Display name is required';

  @override
  String get invalidCredentials => 'Invalid email or password';

  @override
  String get emailInUse => 'Email already registered';

  @override
  String get sessionExpired => 'Session expired. Please log in again.';

  @override
  String get cameraPermissionDenied =>
      'Camera permission is required to scan receipts.';

  @override
  String get openSettings => 'Open Settings';

  @override
  String get noReceiptsYet => 'No receipts yet';

  @override
  String get uploadFirstReceipt => 'Scan your first receipt to start saving.';

  @override
  String get errorOccurred => 'Something went wrong';

  @override
  String get duplicateReceipt => 'This receipt has already been uploaded.';

  @override
  String get viewExisting => 'View Existing';
}
