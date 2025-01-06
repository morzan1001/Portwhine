import 'package:bloc/bloc.dart';

class PipelineSizeCubit extends Cubit<int> {
  PipelineSizeCubit() : super(10);

  void setSize(int size) => emit(size);
}
