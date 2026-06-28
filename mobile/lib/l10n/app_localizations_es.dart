// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appName => 'AhorraApp';

  @override
  String get onboardingTitle1 => 'Encuentra dónde comprar más barato';

  @override
  String get onboardingSubtitle1 =>
      'Escanea tus recibos y descubre los mejores precios cerca de ti.';

  @override
  String get onboardingTitle2 => 'Gana puntos';

  @override
  String get onboardingSubtitle2 =>
      'Recibe recompensas cada vez que confirmas un recibo.';

  @override
  String get onboardingTitle3 => 'Únete a la comunidad';

  @override
  String get onboardingSubtitle3 =>
      'Ayuda a otros a ahorrar compartiendo datos de precios.';

  @override
  String get skip => 'Omitir';

  @override
  String get next => 'Siguiente';

  @override
  String get getStarted => 'Comenzar';

  @override
  String get login => 'Iniciar Sesión';

  @override
  String get register => 'Registrarse';

  @override
  String get email => 'Correo';

  @override
  String get password => 'Contraseña';

  @override
  String get displayName => 'Nombre';

  @override
  String get loginSubtitle =>
      '¡Bienvenido de nuevo! Inicia sesión para continuar.';

  @override
  String get registerSubtitle => 'Crea una cuenta para empezar a ahorrar.';

  @override
  String get noAccount => '¿No tienes cuenta?';

  @override
  String get haveAccount => '¿Ya tienes cuenta?';

  @override
  String get scanReceipt => 'Escanear Recibo';

  @override
  String get uploading => 'Subiendo...';

  @override
  String retrying(Object attempt, Object max) {
    return 'Reintentando ($attempt/$max)...';
  }

  @override
  String get retry => 'Reintentar';

  @override
  String get receiptPending => 'Procesando recibo...';

  @override
  String get receiptNeedsReview => 'Requiere Revisión';

  @override
  String get receiptConfirmed => 'Confirmado';

  @override
  String get receiptRejected => 'Rechazado';

  @override
  String get reviewReceipt => 'Revisar Recibo';

  @override
  String get store => 'Tienda';

  @override
  String get storeName => 'Nombre de Tienda';

  @override
  String get branch => 'Sucursal';

  @override
  String get address => 'Dirección';

  @override
  String get purchaseDate => 'Fecha de Compra';

  @override
  String get total => 'Total';

  @override
  String get items => 'Artículos';

  @override
  String get addItem => 'Agregar Artículo';

  @override
  String get product => 'Producto';

  @override
  String get quantity => 'Cantidad';

  @override
  String get price => 'Precio';

  @override
  String get currency => 'Moneda';

  @override
  String get confirm => 'Confirmar';

  @override
  String get reject => 'Rechazar';

  @override
  String get pointsEarned => 'Puntos Ganados';

  @override
  String get searchProducts => 'Buscar Productos';

  @override
  String get searchHint => 'Buscar un producto...';

  @override
  String get noResults => 'No se encontraron resultados';

  @override
  String noResultsFor(Object query) {
    return 'No se encontraron resultados para \'$query\'';
  }

  @override
  String get cheapestStore => 'Tienda Más Barata';

  @override
  String get profile => 'Perfil';

  @override
  String get totalPoints => 'Puntos Totales';

  @override
  String get recentActivity => 'Actividad Reciente';

  @override
  String get logout => 'Cerrar Sesión';

  @override
  String get home => 'Inicio';

  @override
  String get receipts => 'Recibos';

  @override
  String get currencyRequired =>
      'La moneda es obligatoria para todos los artículos';

  @override
  String get emailRequired => 'El correo es obligatorio';

  @override
  String get passwordRequired => 'La contraseña es obligatoria';

  @override
  String get passwordMinLength =>
      'La contraseña debe tener al menos 8 caracteres';

  @override
  String get nameRequired => 'El nombre es obligatorio';

  @override
  String get invalidCredentials => 'Correo o contraseña incorrectos';

  @override
  String get emailInUse => 'Correo ya registrado';

  @override
  String get sessionExpired =>
      'Sesión expirada. Por favor, inicia sesión de nuevo.';

  @override
  String get cameraPermissionDenied =>
      'Se requiere permiso de cámara para escanear recibos.';

  @override
  String get openSettings => 'Abrir Configuración';

  @override
  String get noReceiptsYet => 'No tienes recibos aún';

  @override
  String get uploadFirstReceipt =>
      'Escanea tu primer recibo para empezar a ahorrar.';

  @override
  String get errorOccurred => 'Algo salió mal';

  @override
  String get duplicateReceipt => 'Este recibo ya ha sido subido.';

  @override
  String get viewExisting => 'Ver Existente';
}
