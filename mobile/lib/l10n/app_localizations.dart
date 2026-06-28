import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_es.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('es'),
  ];

  /// App name
  ///
  /// In en, this message translates to:
  /// **'AhorraApp'**
  String get appName;

  /// First onboarding slide title
  ///
  /// In en, this message translates to:
  /// **'Find where to buy cheaper'**
  String get onboardingTitle1;

  /// First onboarding slide subtitle
  ///
  /// In en, this message translates to:
  /// **'Scan your receipts and discover the best prices near you.'**
  String get onboardingSubtitle1;

  /// Second onboarding slide title
  ///
  /// In en, this message translates to:
  /// **'Earn points'**
  String get onboardingTitle2;

  /// Second onboarding slide subtitle
  ///
  /// In en, this message translates to:
  /// **'Get rewarded every time you confirm a receipt.'**
  String get onboardingSubtitle2;

  /// Third onboarding slide title
  ///
  /// In en, this message translates to:
  /// **'Join the community'**
  String get onboardingTitle3;

  /// Third onboarding slide subtitle
  ///
  /// In en, this message translates to:
  /// **'Help others save money by sharing price data.'**
  String get onboardingSubtitle3;

  /// Skip button
  ///
  /// In en, this message translates to:
  /// **'Skip'**
  String get skip;

  /// Next button
  ///
  /// In en, this message translates to:
  /// **'Next'**
  String get next;

  /// Get started button
  ///
  /// In en, this message translates to:
  /// **'Get Started'**
  String get getStarted;

  /// Login button / title
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get login;

  /// Register button / title
  ///
  /// In en, this message translates to:
  /// **'Register'**
  String get register;

  /// Email field label
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get email;

  /// Password field label
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get password;

  /// Display name field label
  ///
  /// In en, this message translates to:
  /// **'Display Name'**
  String get displayName;

  /// Login subtitle
  ///
  /// In en, this message translates to:
  /// **'Welcome back! Please log in to continue.'**
  String get loginSubtitle;

  /// Register subtitle
  ///
  /// In en, this message translates to:
  /// **'Create an account to start saving.'**
  String get registerSubtitle;

  /// No account prompt
  ///
  /// In en, this message translates to:
  /// **'Don\'t have an account?'**
  String get noAccount;

  /// Have account prompt
  ///
  /// In en, this message translates to:
  /// **'Already have an account?'**
  String get haveAccount;

  /// Scan receipt button
  ///
  /// In en, this message translates to:
  /// **'Scan Receipt'**
  String get scanReceipt;

  /// Upload in progress
  ///
  /// In en, this message translates to:
  /// **'Uploading...'**
  String get uploading;

  /// Retry in progress with attempt count
  ///
  /// In en, this message translates to:
  /// **'Retrying ({attempt}/{max})...'**
  String retrying(Object attempt, Object max);

  /// Retry button
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// Receipt pending status
  ///
  /// In en, this message translates to:
  /// **'Processing receipt...'**
  String get receiptPending;

  /// Receipt needs review status
  ///
  /// In en, this message translates to:
  /// **'Needs Review'**
  String get receiptNeedsReview;

  /// Receipt confirmed status
  ///
  /// In en, this message translates to:
  /// **'Confirmed'**
  String get receiptConfirmed;

  /// Receipt rejected status
  ///
  /// In en, this message translates to:
  /// **'Rejected'**
  String get receiptRejected;

  /// Review receipt title
  ///
  /// In en, this message translates to:
  /// **'Review Receipt'**
  String get reviewReceipt;

  /// Store label
  ///
  /// In en, this message translates to:
  /// **'Store'**
  String get store;

  /// Store name field
  ///
  /// In en, this message translates to:
  /// **'Store Name'**
  String get storeName;

  /// Branch field
  ///
  /// In en, this message translates to:
  /// **'Branch'**
  String get branch;

  /// Address field
  ///
  /// In en, this message translates to:
  /// **'Address'**
  String get address;

  /// Purchase date label
  ///
  /// In en, this message translates to:
  /// **'Purchase Date'**
  String get purchaseDate;

  /// Total label
  ///
  /// In en, this message translates to:
  /// **'Total'**
  String get total;

  /// Items label
  ///
  /// In en, this message translates to:
  /// **'Items'**
  String get items;

  /// Add item button
  ///
  /// In en, this message translates to:
  /// **'Add Item'**
  String get addItem;

  /// Product field
  ///
  /// In en, this message translates to:
  /// **'Product'**
  String get product;

  /// Quantity field
  ///
  /// In en, this message translates to:
  /// **'Quantity'**
  String get quantity;

  /// Price field
  ///
  /// In en, this message translates to:
  /// **'Price'**
  String get price;

  /// Currency label
  ///
  /// In en, this message translates to:
  /// **'Currency'**
  String get currency;

  /// Confirm button
  ///
  /// In en, this message translates to:
  /// **'Confirm'**
  String get confirm;

  /// Reject button
  ///
  /// In en, this message translates to:
  /// **'Reject'**
  String get reject;

  /// Points earned title
  ///
  /// In en, this message translates to:
  /// **'Points Earned'**
  String get pointsEarned;

  /// Search products title
  ///
  /// In en, this message translates to:
  /// **'Search Products'**
  String get searchProducts;

  /// Search hint
  ///
  /// In en, this message translates to:
  /// **'Search for a product...'**
  String get searchHint;

  /// No results message
  ///
  /// In en, this message translates to:
  /// **'No results found'**
  String get noResults;

  /// No results for query
  ///
  /// In en, this message translates to:
  /// **'No results found for \'{query}\''**
  String noResultsFor(Object query);

  /// Cheapest store label
  ///
  /// In en, this message translates to:
  /// **'Cheapest Store'**
  String get cheapestStore;

  /// Profile title
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// Total points label
  ///
  /// In en, this message translates to:
  /// **'Total Points'**
  String get totalPoints;

  /// Recent activity label
  ///
  /// In en, this message translates to:
  /// **'Recent Activity'**
  String get recentActivity;

  /// Logout button
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// Home tab
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get home;

  /// Receipts tab
  ///
  /// In en, this message translates to:
  /// **'Receipts'**
  String get receipts;

  /// Currency required error
  ///
  /// In en, this message translates to:
  /// **'Currency is required for all items'**
  String get currencyRequired;

  /// Email required error
  ///
  /// In en, this message translates to:
  /// **'Email is required'**
  String get emailRequired;

  /// Password required error
  ///
  /// In en, this message translates to:
  /// **'Password is required'**
  String get passwordRequired;

  /// Password min length error
  ///
  /// In en, this message translates to:
  /// **'Password must be at least 8 characters'**
  String get passwordMinLength;

  /// Name required error
  ///
  /// In en, this message translates to:
  /// **'Display name is required'**
  String get nameRequired;

  /// Invalid credentials error
  ///
  /// In en, this message translates to:
  /// **'Invalid email or password'**
  String get invalidCredentials;

  /// Email in use error
  ///
  /// In en, this message translates to:
  /// **'Email already registered'**
  String get emailInUse;

  /// Session expired message
  ///
  /// In en, this message translates to:
  /// **'Session expired. Please log in again.'**
  String get sessionExpired;

  /// Camera permission denied
  ///
  /// In en, this message translates to:
  /// **'Camera permission is required to scan receipts.'**
  String get cameraPermissionDenied;

  /// Open settings button
  ///
  /// In en, this message translates to:
  /// **'Open Settings'**
  String get openSettings;

  /// No receipts empty state
  ///
  /// In en, this message translates to:
  /// **'No receipts yet'**
  String get noReceiptsYet;

  /// Upload first receipt prompt
  ///
  /// In en, this message translates to:
  /// **'Scan your first receipt to start saving.'**
  String get uploadFirstReceipt;

  /// Generic error message
  ///
  /// In en, this message translates to:
  /// **'Something went wrong'**
  String get errorOccurred;

  /// Duplicate receipt message
  ///
  /// In en, this message translates to:
  /// **'This receipt has already been uploaded.'**
  String get duplicateReceipt;

  /// View existing receipt button
  ///
  /// In en, this message translates to:
  /// **'View Existing'**
  String get viewExisting;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'es'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'es':
      return AppLocalizationsEs();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
