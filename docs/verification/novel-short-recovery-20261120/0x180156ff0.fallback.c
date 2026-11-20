/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180156ff0 */

longlong * FUN_180156ff0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong lVar2;
  int *piVar3;
  longlong *plVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong lVar12;
  undefined8 uVar13;
  code *pcVar14;
  code *pcVar15;

  plVar4 = DAT_1803b6b28;
  plVar1 = (longlong *)*param_3;
  if (DAT_1803b6b28 == (longlong *)0x0) {
LAB_18015704e:
    DAT_1803b6b28 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b98,DAT_1803ca758,8);
  }
  else if ((1 < *DAT_1803b6b28) || (DAT_1803b6b28[2] != 0)) {
    if ((DAT_1803b6b28 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b28 && (*DAT_1803b6b28 = *DAT_1803b6b28 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_18015704e;
  }
  plVar4 = DAT_1803b6b28;
  FUN_180001d00(param_1,DAT_1803b6b28);
  plVar7 = (longlong *)FUN_18037f0c0();
  if (plVar7 == (longlong *)0x0) {
    plVar10 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    plVar8 = (longlong *)FUN_18000d280();
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) {
      plVar10 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar9 = (longlong *)FUN_180372160(param_1,0);
      plVar7 = (longlong *)0x0;
      do {
        pcVar14 = *(code **)(plVar8[1] + 0xe0);
        if (pcVar14 == (code *)0x0) {
          PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                       *(undefined8 *)(plVar8[1] + 0x18));
LAB_1801573a4:
          plVar10 = *(longlong **)(param_1 + 0x60);
          if (plVar10 == (longlong *)0x0) {
LAB_1801575d1:
            iVar6 = (int)*plVar9 + 1;
            if (iVar6 != 0) {
              *(int *)plVar9 = iVar6;
            }
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            FUN_180001d80(param_1);
            if ((int)*plVar1 < 0) {
              return plVar9;
            }
            *plVar1 = *plVar1 + -1;
            if (*plVar1 != 0) {
              return plVar9;
            }
            (**(code **)(plVar1[1] + 0x30))();
            return plVar9;
          }
          plVar11 = (longlong *)plVar10[1];
          if ((*(uint *)((longlong *)plVar10[1] + 0x15) & 0x40000000) == 0) {
            plVar11 = plVar10;
          }
          if ((plVar11 == *(longlong **)PyExc_StopIteration_exref) ||
             (((*(int *)(plVar11[1] + 0xa8) < 0 && ((*(uint *)(plVar11 + 0x15) & 0x40000000) != 0))
              && (cVar5 = FUN_180382640(), cVar5 == '\x01')))) {
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            goto LAB_1801575d1;
          }
LAB_1801573fc:
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_180157404;
        }
        plVar10 = (longlong *)(*pcVar14)(plVar8);
        if (plVar10 == (longlong *)0x0) goto LAB_1801573a4;
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        iVar6 = (int)*plVar10 + 1;
        if (iVar6 != 0) {
          *(int *)plVar10 = iVar6;
        }
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        plVar11 = (longlong *)FUN_18037f0c0();
        plVar7 = plVar10;
        if (plVar11 == (longlong *)0x0) {
LAB_180157325:
          plVar10 = *(longlong **)(param_1 + 0x60);
          goto LAB_1801573fc;
        }
        plVar10 = (longlong *)FUN_18037f0c0();
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (plVar10 == (longlong *)0x0) goto LAB_180157325;
        *(undefined4 *)(plVar4 + 5) = 0xf7;
        pcVar14 = (code *)FUN_180376210(param_1,plVar10);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (pcVar14 == (code *)0x0) goto LAB_180157325;
        if (pcVar14 == _Py_TrueStruct_exref) {
LAB_1801571f3:
          iVar6 = 1;
        }
        else if ((pcVar14 == _Py_FalseStruct_exref) || (pcVar14 == _Py_NoneStruct_exref)) {
          iVar6 = 0;
        }
        else {
          lVar12 = *(longlong *)(pcVar14 + 8);
          if ((*(longlong *)(lVar12 + 0x60) == 0) ||
             (pcVar15 = *(code **)(*(longlong *)(lVar12 + 0x60) + 0x48), pcVar15 == (code *)0x0)) {
            if ((*(undefined8 **)(lVar12 + 0x70) == (undefined8 *)0x0) ||
               (pcVar15 = (code *)**(undefined8 **)(lVar12 + 0x70), pcVar15 == (code *)0x0)) {
              if ((*(undefined8 **)(lVar12 + 0x68) == (undefined8 *)0x0) ||
                 (pcVar15 = (code *)**(undefined8 **)(lVar12 + 0x68), pcVar15 == (code *)0x0))
              goto LAB_1801571f3;
              lVar12 = (*pcVar15)(pcVar14);
            }
            else {
              lVar12 = (*pcVar15)(pcVar14);
            }
          }
          else {
            iVar6 = (*pcVar15)(pcVar14);
            lVar12 = (longlong)iVar6;
          }
          if (lVar12 < 1) {
            iVar6 = -(uint)(lVar12 != 0);
          }
          else {
            iVar6 = 1;
          }
        }
        if ((-1 < *(int *)pcVar14) &&
           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
        }
        if (iVar6 == -1) goto LAB_180157325;
        if (iVar6 == 0) {
          lVar12 = FUN_18037f0c0();
          if (lVar12 != 0) {
            lVar2 = plVar9[2];
            cVar5 = FUN_1803722b0(plVar9,lVar2 + 1);
            if (cVar5 != '\0') {
              *(longlong *)(plVar9[3] + lVar2 * 8) = lVar12;
              goto LAB_1801572c6;
            }
          }
          goto LAB_180157325;
        }
LAB_1801572c6:
        lVar12 = *(longlong *)(param_1 + 0x10);
        if ((((*(int *)(*(longlong *)(lVar12 + 0x60) + 0x4d4) != 0) ||
             (*(int *)(lVar12 + 0x1a0) != 0)) && (iVar6 = Py_MakePendingCalls(), iVar6 < 0)) &&
           (*(longlong *)(param_1 + 0x60) != 0)) {
          plVar10 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_180157404;
        }
        if (*(int *)(lVar12 + 0x174) != 0) {
          PyEval_SaveThread();
          PyEval_AcquireThread(param_1);
        }
      } while (*(longlong *)(param_1 + 0x80) == 0);
      plVar10 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x80) = 0;
      uVar13 = FUN_180373490(param_1);
      *(undefined8 *)(param_1 + 0x60) = uVar13;
      if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      plVar10 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_180157404:
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))();
      }
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
    }
  }
  piVar3 = (int *)plVar10[5];
  if (piVar3 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar4,0xf7);
    plVar7 = (longlong *)plVar10[5];
    pcVar15 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
LAB_1801574e0:
      pcVar14 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_180157529;
    pcVar14 = (code *)FUN_18038e2a0(plVar4,0xf7);
    *(int **)(pcVar14 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7 = (longlong *)plVar10[5];
    pcVar15 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_1801574e0;
    pcVar15 = (code *)0x0;
  }
  plVar10[5] = (longlong)pcVar15;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_180157529:
  FUN_18036d0a0(plVar4,&DAT_180394940,plVar1);
  plVar7 = DAT_1803b6b28;
  if (plVar4 == DAT_1803b6b28) {
    if ((-1 < (int)*DAT_1803b6b28) && (*DAT_1803b6b28 = *DAT_1803b6b28 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b6b28 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar10;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
