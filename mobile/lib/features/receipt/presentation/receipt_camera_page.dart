import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';

import '../../../app/theme/spacing.dart';
import 'receipt_upload_notifier.dart';

class ReceiptCameraPage extends ConsumerStatefulWidget {
  const ReceiptCameraPage({super.key});

  static const routeName = 'receipt-camera';

  @override
  ConsumerState<ReceiptCameraPage> createState() => _ReceiptCameraPageState();
}

enum _CameraPhase { initializing, ready, denied, error, uploading }

class _ReceiptCameraPageState extends ConsumerState<ReceiptCameraPage>
    with WidgetsBindingObserver {
  CameraController? _controller;
  ProviderSubscription<UploadState>? _uploadSubscription;
  _CameraPhase _phase = _CameraPhase.initializing;
  String? _errorMessage;
  bool _flashOn = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _initializeCamera();
    _uploadSubscription = ref.listenManual<UploadState>(
      uploadNotifierProvider,
      (previous, next) {
        if (!mounted) return;
        switch (next.status) {
          case UploadStatus.success:
            if (next.receiptId != null) {
              context.go('/receipts/${next.receiptId}');
            }
          case UploadStatus.duplicate:
            if (next.receiptId != null) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('This receipt has already been uploaded.'),
                ),
              );
              context.go('/receipts/${next.receiptId}');
            }
          case UploadStatus.error:
            setState(() => _phase = _CameraPhase.ready);
            if (next.errorMessage != null) {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(SnackBar(content: Text(next.errorMessage!)));
            }
          case UploadStatus.idle:
          case UploadStatus.uploading:
            break;
        }
      },
    );
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _uploadSubscription?.close();
    _controller?.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    final controller = _controller;
    if (controller == null || !controller.value.isInitialized) return;

    if (state == AppLifecycleState.inactive) {
      controller.dispose();
    } else if (state == AppLifecycleState.resumed) {
      _initializeCamera();
    }
  }

  Future<void> _initializeCamera() async {
    try {
      final cameras = await availableCameras();
      if (cameras.isEmpty) {
        setState(() {
          _phase = _CameraPhase.denied;
          _errorMessage = 'No camera available on this device.';
        });
        return;
      }
      final back = cameras.firstWhere(
        (c) => c.lensDirection == CameraLensDirection.back,
        orElse: () => cameras.first,
      );
      final controller = CameraController(
        back,
        ResolutionPreset.high,
        enableAudio: false,
      );
      await controller.initialize();
      if (!mounted) {
        await controller.dispose();
        return;
      }
      setState(() {
        _controller = controller;
        _phase = _CameraPhase.ready;
        _errorMessage = null;
      });
    } on CameraException catch (e) {
      final code = e.code;
      final isPermissionDenial =
          code == 'CameraAccessDenied' ||
          code == 'CameraAccessDeniedWithoutPrompt';
      setState(() {
        _phase = isPermissionDenial ? _CameraPhase.denied : _CameraPhase.error;
        _errorMessage = e.description?.isNotEmpty == true
            ? e.description
            : 'Camera error ($code).';
      });
    } on Exception catch (e) {
      setState(() {
        _phase = _CameraPhase.error;
        _errorMessage = e.toString();
      });
    }
  }

  Future<void> _capturePhoto() async {
    final controller = _controller;
    if (controller == null || !controller.value.isInitialized) return;
    XFile file;
    try {
      file = await controller.takePicture();
    } on CameraException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.description ?? 'Capture failed.')),
      );
      return;
    }
    if (!mounted) return;
    setState(() => _phase = _CameraPhase.uploading);
    final bytes = await file.readAsBytes();
    await ref
        .read(uploadNotifierProvider.notifier)
        .uploadBytes(bytes, filename: file.name);
  }

  Future<void> _pickFromGallery() async {
    final picker = ImagePicker();
    final photo = await picker.pickImage(source: ImageSource.gallery);
    if (photo == null) return;
    setState(() => _phase = _CameraPhase.uploading);
    final bytes = await photo.readAsBytes();
    await ref
        .read(uploadNotifierProvider.notifier)
        .uploadBytes(bytes, filename: photo.name);
  }

  Future<void> _toggleFlash() async {
    final controller = _controller;
    if (controller == null || !controller.value.isInitialized) return;
    final next = !_flashOn;
    try {
      await controller.setFlashMode(next ? FlashMode.torch : FlashMode.off);
      setState(() => _flashOn = next);
    } on CameraException {
      // Some devices do not support torch; silently ignore.
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          fit: StackFit.expand,
          children: [
            _buildPreview(),
            if (_phase == _CameraPhase.ready ||
                _phase == _CameraPhase.uploading)
              _buildOverlay(),
            Positioned(top: 0, left: 0, right: 0, child: _buildTopBar()),
            Positioned(bottom: 0, left: 0, right: 0, child: _buildControls()),
          ],
        ),
      ),
    );
  }

  Widget _buildPreview() {
    switch (_phase) {
      case _CameraPhase.initializing:
        return const ColoredBox(
          color: Colors.black,
          child: Center(child: CircularProgressIndicator(color: Colors.white)),
        );
      case _CameraPhase.denied:
        return _buildDenied();
      case _CameraPhase.error:
        return _buildError();
      case _CameraPhase.ready:
      case _CameraPhase.uploading:
        final controller = _controller;
        if (controller == null || !controller.value.isInitialized) {
          return const Center(
            child: CircularProgressIndicator(color: Colors.white),
          );
        }
        return CameraPreview(controller);
    }
  }

  Widget _buildOverlay() {
    return IgnorePointer(
      ignoring: _phase == _CameraPhase.uploading,
      child: CustomPaint(
        painter: _FramingOverlayPainter(),
        child: const Align(
          alignment: Alignment(0, -0.25),
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: AppSpacing.xl),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Align the receipt within the frame',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                SizedBox(height: AppSpacing.xs),
                Text(
                  'Keep it flat and well lit for accurate OCR',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Colors.white70, fontSize: 12),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTopBar() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            IconButton(
              icon: const Icon(Icons.close, color: Colors.white),
              tooltip: 'Close',
              onPressed: () => Navigator.of(context).maybePop(),
            ),
            if (_phase == _CameraPhase.ready)
              IconButton(
                icon: Icon(
                  _flashOn ? Icons.flash_on : Icons.flash_off,
                  color: Colors.white,
                ),
                tooltip: 'Toggle flash',
                onPressed: _toggleFlash,
              )
            else
              const SizedBox(width: 48),
          ],
        ),
      ),
    );
  }

  Widget _buildControls() {
    if (_phase == _CameraPhase.uploading) {
      return const SafeArea(
        child: Padding(
          padding: EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              LinearProgressIndicator(color: Colors.white),
              SizedBox(height: AppSpacing.sm),
              Text('Uploading...', style: TextStyle(color: Colors.white)),
            ],
          ),
        ),
      );
    }
    if (_phase != _CameraPhase.ready) {
      return const SizedBox();
    }
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            IconButton(
              icon: const Icon(Icons.photo_outlined, color: Colors.white),
              tooltip: 'Pick from gallery',
              onPressed: _pickFromGallery,
            ),
            _CaptureButton(onTap: _capturePhoto),
            const SizedBox(width: 48),
          ],
        ),
      ),
    );
  }

  Widget _buildDenied() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.no_photography, size: 72, color: Colors.white70),
            const SizedBox(height: AppSpacing.lg),
            const Text(
              'Camera permission is required to scan receipts.',
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.white, fontSize: 16),
            ),
            const SizedBox(height: AppSpacing.md),
            Text(
              _errorMessage ?? 'Enable camera access in your device settings.',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.white70),
            ),
            const SizedBox(height: AppSpacing.xl),
            FilledButton.icon(
              onPressed: _initializeCamera,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
            const SizedBox(height: AppSpacing.md),
            OutlinedButton.icon(
              onPressed: _pickFromGallery,
              icon: const Icon(Icons.photo_outlined),
              label: const Text('Pick from gallery instead'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildError() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.error_outline, size: 72, color: Colors.white70),
            const SizedBox(height: AppSpacing.lg),
            Text(
              _errorMessage ?? 'Camera failed to start.',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.white),
            ),
            const SizedBox(height: AppSpacing.xl),
            FilledButton.icon(
              onPressed: _initializeCamera,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
            const SizedBox(height: AppSpacing.md),
            OutlinedButton.icon(
              onPressed: _pickFromGallery,
              icon: const Icon(Icons.photo_outlined),
              label: const Text('Pick from gallery'),
            ),
          ],
        ),
      ),
    );
  }
}

class _CaptureButton extends StatelessWidget {
  const _CaptureButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: 'Capture receipt photo',
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(40),
        child: Container(
          width: 72,
          height: 72,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white, width: 4),
          ),
          alignment: Alignment.center,
          child: Container(
            width: 56,
            height: 56,
            decoration: const BoxDecoration(
              color: Colors.white,
              shape: BoxShape.circle,
            ),
          ),
        ),
      ),
    );
  }
}

class _FramingOverlayPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    const radius = AppRadius.lg;
    final opening = Rect.fromCenter(
      center: Offset(size.width / 2, size.height / 2),
      width: size.width * 0.78,
      height: size.height * 0.62,
    );

    final rrect = RRect.fromRectAndRadius(
      opening,
      const Radius.circular(radius),
    );

    final paint = Paint()..color = Colors.black.withValues(alpha: 0.5);
    canvas.saveLayer(Offset.zero & size, paint);
    canvas.drawRect(Offset.zero & size, paint);
    paint.blendMode = BlendMode.clear;
    canvas.drawRRect(rrect, paint);
    canvas.restore();

    final border = Paint()
      ..color = Colors.white.withValues(alpha: 0.85)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5;
    canvas.drawRRect(rrect, border);

    final bracketPaint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3.0
      ..strokeCap = StrokeCap.round;
    const bracketLen = 28.0;
    final corners = <Path>[
      Path()
        ..moveTo(opening.left, opening.top + bracketLen)
        ..lineTo(opening.left, opening.top)
        ..lineTo(opening.left + bracketLen, opening.top),
      Path()
        ..moveTo(opening.right - bracketLen, opening.top)
        ..lineTo(opening.right, opening.top)
        ..lineTo(opening.right, opening.top + bracketLen),
      Path()
        ..moveTo(opening.left, opening.bottom - bracketLen)
        ..lineTo(opening.left, opening.bottom)
        ..lineTo(opening.left + bracketLen, opening.bottom),
      Path()
        ..moveTo(opening.right - bracketLen, opening.bottom)
        ..lineTo(opening.right, opening.bottom)
        ..lineTo(opening.right, opening.bottom - bracketLen),
    ];
    for (final path in corners) {
      canvas.drawPath(path, bracketPaint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
