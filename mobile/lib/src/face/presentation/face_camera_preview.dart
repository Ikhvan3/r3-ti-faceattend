import 'package:camera/camera.dart';
import 'package:flutter/material.dart';

class FaceCameraPreview extends StatelessWidget {
  const FaceCameraPreview({required this.controller, super.key});

  final CameraController controller;

  @override
  Widget build(BuildContext context) {
    final mediaOrientation = MediaQuery.of(context).orientation;
    final isPortrait = mediaOrientation == Orientation.portrait;
    final previewAspectRatio = mediaOrientation == Orientation.portrait
        ? 1 / controller.value.aspectRatio
        : controller.value.aspectRatio;
    final previewSize = controller.value.previewSize;
    final previewWidth = previewSize == null
        ? 1.0
        : isPortrait
        ? previewSize.height
        : previewSize.width;
    final previewHeight = previewSize == null
        ? 1.0
        : isPortrait
        ? previewSize.width
        : previewSize.height;

    return AspectRatio(
      aspectRatio: previewAspectRatio,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: FittedBox(
          fit: BoxFit.cover,
          child: SizedBox(
            width: previewWidth,
            height: previewHeight,
            child: CameraPreview(controller),
          ),
        ),
      ),
    );
  }
}
