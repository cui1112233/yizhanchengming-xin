/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801519b0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

code * FUN_1801519b0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  code *pcVar2;
  code *pcVar3;
  char cVar4;
  int iVar5;
  longlong *plVar6;
  longlong *plVar7;
  code *pcVar8;
  longlong lVar9;
  longlong *plVar10;
  undefined8 uVar11;
  code *pcVar12;
  code *pcVar13;
  code *pcVar14;
  uint uVar15;
  longlong *plVar16;
  longlong lVar17;
  undefined8 uVar18;
  undefined1 auStack_c8 [32];
  code *local_a8;
  code *local_90;
  code *local_88;
  code *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  undefined8 local_60;
  code *local_58;
  code *local_50;
  ulonglong local_48;

  plVar16 = DAT_1803b6a78;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_c8;
  pcVar12 = (code *)0x0;
  local_68 = (longlong *)*param_3;
  local_88 = (code *)0x0;
  local_90 = (code *)0x0;
  local_80 = (code *)0x0;
  local_78 = DAT_180391d20;
  local_70 = DAT_1803b6a78;
  if (DAT_1803b6a78 == (longlong *)0x0) {
LAB_180151a3e:
    DAT_1803b6a78 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6a70,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6a78) || (DAT_1803b6a78[2] != 0)) {
    if ((DAT_1803b6a78 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6a78 && (*DAT_1803b6a78 = *DAT_1803b6a78 + -1, *plVar16 == 0)))) {
      (**(code **)(plVar16[1] + 0x30))(plVar16);
    }
    goto LAB_180151a3e;
  }
  plVar16 = DAT_1803b6a78;
  local_70 = DAT_1803b6a78;
  FUN_180001d00(param_1,DAT_1803b6a78);
  plVar6 = (longlong *)FUN_18037f0c0();
  uVar18 = 0x7e;
  if (plVar6 == (longlong *)0x0) {
    plVar7 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *(undefined4 *)(plVar16 + 5) = 0x7e;
    plVar7 = (longlong *)FUN_180376210(param_1,plVar6);
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    if (plVar7 == (longlong *)0x0) {
      plVar7 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar12 = (code *)0x0;
    }
    else {
      plVar6 = (longlong *)FUN_18000d280();
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      pcVar13 = pcVar12;
      pcVar14 = pcVar12;
      if (plVar6 == (longlong *)0x0) {
        plVar7 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar16 = local_70;
      }
      else {
        do {
          pcVar2 = local_80;
          pcVar3 = local_88;
          pcVar12 = *(code **)(plVar6[1] + 0xe0);
          if (pcVar12 == (code *)0x0) {
            PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                         *(undefined8 *)(plVar6[1] + 0x18));
            pcVar12 = pcVar14;
LAB_18015200e:
            plVar7 = *(longlong **)(param_1 + 0x60);
            if (plVar7 != (longlong *)0x0) {
              plVar16 = (longlong *)plVar7[1];
              if ((*(uint *)((longlong *)plVar7[1] + 0x15) & 0x40000000) == 0) {
                plVar16 = plVar7;
              }
              if ((plVar16 != *(longlong **)PyExc_StopIteration_exref) &&
                 (((-1 < *(int *)(plVar16[1] + 0xa8) ||
                   ((*(uint *)(plVar16 + 0x15) & 0x40000000) == 0)) ||
                  (cVar4 = FUN_180382640(), cVar4 != '\x01')))) {
                *(undefined8 *)(param_1 + 0x60) = 0;
                local_88 = pcVar3;
                local_80 = pcVar13;
                if (pcVar13 != (code *)0x0) goto LAB_18015207d;
                goto LAB_18015209c;
              }
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
            }
            FUN_180001d80(param_1);
            if (((local_80 != (code *)0x0) && (-1 < *(int *)local_80)) &&
               (*(longlong *)local_80 = *(longlong *)local_80 + -1, *(longlong *)local_80 == 0)) {
              (**(code **)(*(longlong *)(local_80 + 8) + 0x30))();
            }
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            pcVar12 = DAT_1803b6c50;
            if (*(int *)DAT_1803b6c50 + 1 != 0) {
              *(int *)DAT_1803b6c50 = *(int *)DAT_1803b6c50 + 1;
            }
            pcVar8 = local_88;
            pcVar13 = local_90;
            if (local_88 == (code *)0x0) goto LAB_1801522d2;
LAB_1801522bd:
            if ((-1 < *(int *)pcVar8) &&
               (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
              (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))();
            }
LAB_1801522d2:
            if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
               (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
              (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
            }
            if ((int)*local_68 < 0) {
              return pcVar12;
            }
            *local_68 = *local_68 + -1;
            if (*local_68 != 0) {
              return pcVar12;
            }
            (**(code **)(local_68[1] + 0x30))(local_68);
            return pcVar12;
          }
          pcVar8 = (code *)(*pcVar12)(plVar6);
          pcVar12 = local_90;
          if (pcVar8 == (code *)0x0) goto LAB_18015200e;
          local_80 = pcVar8;
          if (((pcVar2 != (code *)0x0) && (-1 < *(int *)pcVar2)) &&
             (*(longlong *)pcVar2 = *(longlong *)pcVar2 + -1, *(longlong *)pcVar2 == 0)) {
            (**(code **)(*(longlong *)(pcVar2 + 8) + 0x30))(pcVar2);
          }
          if (*(int *)pcVar8 + 1 != 0) {
            *(int *)pcVar8 = *(int *)pcVar8 + 1;
          }
          local_88 = pcVar8;
          if (((pcVar3 != (code *)0x0) && (-1 < *(int *)pcVar3)) &&
             (*(longlong *)pcVar3 = *(longlong *)pcVar3 + -1, *(longlong *)pcVar3 == 0)) {
            (**(code **)(*(longlong *)(pcVar3 + 8) + 0x30))();
          }
          lVar9 = FUN_18014c500(param_1);
          plVar16 = local_70;
          if (lVar9 == 0) {
            FUN_180373340(param_1,&local_78,DAT_1803b6c30);
            uVar18 = 0x7f;
            plVar7 = local_78;
            pcVar12 = local_90;
            goto LAB_18015207d;
          }
          local_60 = DAT_1803b6d40;
          local_58 = DAT_1803b6c50;
          *(undefined4 *)(local_70 + 5) = 0x7f;
          local_50 = pcVar8;
          plVar7 = (longlong *)FUN_18037e980(param_1,lVar9,DAT_1803b6d18,&local_60);
          if (plVar7 == (longlong *)0x0) {
LAB_180151fa7:
            plVar7 = *(longlong **)(param_1 + 0x60);
            uVar18 = 0x7f;
            *(undefined8 *)(param_1 + 0x60) = 0;
            pcVar12 = local_90;
            goto LAB_18015207d;
          }
          lVar9 = plVar7[1];
          if (*(code **)(lVar9 + 0x90) == (code *)0x0) {
            uVar15 = *(uint *)(DAT_1803b6cb8 + 0x20) & 0x60;
            if (*(code **)(lVar9 + 0x40) == (code *)0x0) {
              lVar17 = DAT_1803b6cb8 + 0x28;
              if (uVar15 != 0x60) {
                lVar17 = *(longlong *)(DAT_1803b6cb8 + 0x30);
              }
              PyErr_Format(*(undefined8 *)PyExc_AttributeError_exref,
                           "\'%s\' object has no attribute \'%s\'",*(undefined8 *)(lVar9 + 0x18),
                           lVar17);
              plVar10 = (longlong *)0x0;
            }
            else {
              lVar17 = DAT_1803b6cb8 + 0x28;
              if (uVar15 != 0x60) {
                lVar17 = *(longlong *)(DAT_1803b6cb8 + 0x30);
              }
              plVar10 = (longlong *)(**(code **)(lVar9 + 0x40))(plVar7,lVar17);
            }
          }
          else {
            plVar10 = (longlong *)(**(code **)(lVar9 + 0x90))(plVar7,DAT_1803b6cb8);
          }
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if (plVar10 == (longlong *)0x0) goto LAB_180151fa7;
          *(undefined4 *)(plVar16 + 5) = 0x7f;
          pcVar13 = (code *)FUN_180376210(param_1,plVar10);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          pcVar12 = local_90;
          if (pcVar13 == (code *)0x0) goto LAB_180151fa7;
          pcVar3 = pcVar13;
          if (((local_90 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
             (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
            pcVar14 = local_90 + 8;
            local_90 = pcVar13;
            (**(code **)(*(longlong *)pcVar14 + 0x30))(pcVar12);
            pcVar3 = local_90;
          }
          local_90 = pcVar3;
          if (pcVar13 == _Py_TrueStruct_exref) {
LAB_180151e49:
            lVar9 = FUN_18014c500(param_1);
            plVar16 = local_70;
            if (lVar9 == 0) {
              FUN_180373340(param_1,&local_78,DAT_1803b6c30);
              uVar18 = 0x81;
              plVar7 = local_78;
              pcVar12 = local_90;
            }
            else {
              uVar18 = 0x81;
              local_58 = DAT_1803b6d50;
              local_60 = DAT_1803b6d48;
              *(undefined4 *)(local_70 + 5) = 0x81;
              local_50 = pcVar13;
              plVar7 = (longlong *)FUN_18037e980(param_1,lVar9,DAT_1803b6d18,&local_60);
              if (plVar7 != (longlong *)0x0) {
                plVar10 = (longlong *)FUN_18037f0c0();
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                if (plVar10 == (longlong *)0x0) goto LAB_180151eca;
                *(undefined4 *)(plVar16 + 5) = 0x81;
                pcVar12 = (code *)FUN_180376210(param_1,plVar10);
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (pcVar12 == (code *)0x0) goto LAB_180151eca;
                if ((-1 < *(int *)pcVar13) &&
                   (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
                  (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                }
                if (*(int *)pcVar12 + 1 != 0) {
                  *(int *)pcVar12 = *(int *)pcVar12 + 1;
                }
                if ((-1 < *(int *)pcVar8) &&
                   (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
                  (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
                }
                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                }
                FUN_180001d80(param_1);
                pcVar13 = pcVar12;
                goto LAB_1801522bd;
              }
LAB_180151eca:
              plVar7 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              pcVar12 = local_90;
            }
            goto LAB_18015207d;
          }
          if ((pcVar13 != _Py_FalseStruct_exref) && (pcVar13 != _Py_NoneStruct_exref)) {
            lVar9 = *(longlong *)(pcVar13 + 8);
            if ((*(longlong *)(lVar9 + 0x60) == 0) ||
               (pcVar12 = *(code **)(*(longlong *)(lVar9 + 0x60) + 0x48), pcVar12 == (code *)0x0)) {
              if ((*(undefined8 **)(lVar9 + 0x70) == (undefined8 *)0x0) ||
                 (pcVar12 = (code *)**(undefined8 **)(lVar9 + 0x70), pcVar12 == (code *)0x0)) {
                if ((*(undefined8 **)(lVar9 + 0x68) == (undefined8 *)0x0) ||
                   (pcVar12 = (code *)**(undefined8 **)(lVar9 + 0x68), pcVar12 == (code *)0x0))
                goto LAB_180151e49;
                lVar9 = (*pcVar12)(pcVar13);
              }
              else {
                lVar9 = (*pcVar12)(pcVar13);
              }
            }
            else {
              iVar5 = (*pcVar12)(pcVar13);
              lVar9 = (longlong)iVar5;
            }
            if (0 < lVar9) goto LAB_180151e49;
            if (lVar9 != 0) {
              plVar7 = *(longlong **)(param_1 + 0x60);
              uVar18 = 0x80;
              *(undefined8 *)(param_1 + 0x60) = 0;
              pcVar12 = local_90;
              goto LAB_18015207d;
            }
          }
          lVar9 = *(longlong *)(param_1 + 0x10);
          if (((*(int *)(*(longlong *)(lVar9 + 0x60) + 0x4d4) != 0) ||
              (*(int *)(lVar9 + 0x1a0) != 0)) &&
             ((iVar5 = Py_MakePendingCalls(), iVar5 < 0 && (*(longlong *)(param_1 + 0x60) != 0))))
          goto LAB_180151e35;
          if (*(int *)(lVar9 + 0x174) != 0) {
            PyEval_SaveThread();
            PyEval_AcquireThread(param_1);
          }
          pcVar13 = local_80;
          pcVar14 = local_90;
        } while (*(longlong *)(param_1 + 0x80) == 0);
        plVar16 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x80) = 0;
        uVar11 = FUN_180373490(param_1);
        *(undefined8 *)(param_1 + 0x60) = uVar11;
        if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
           (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
          (**(code **)(plVar16[1] + 0x30))(plVar16);
        }
LAB_180151e35:
        plVar7 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        pcVar12 = local_90;
LAB_18015207d:
        if ((-1 < *(int *)local_80) &&
           (*(longlong *)local_80 = *(longlong *)local_80 + -1, *(longlong *)local_80 == 0)) {
          (**(code **)(*(longlong *)(local_80 + 8) + 0x30))(local_80);
        }
LAB_18015209c:
        plVar16 = local_70;
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
          plVar16 = local_70;
        }
      }
    }
  }
  piVar1 = (int *)plVar7[5];
  if (piVar1 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar16,uVar18);
    plVar6 = (longlong *)plVar7[5];
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_18015211e:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar16) goto LAB_180152167;
    pcVar13 = (code *)FUN_18038e2a0(plVar16,uVar18);
    *(int **)(pcVar13 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar6 = (longlong *)plVar7[5];
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_18015211e;
    pcVar14 = (code *)0x0;
  }
  plVar7[5] = (longlong)pcVar14;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_180152167:
  plVar10 = local_68;
  pcVar13 = local_88;
  local_a8 = pcVar12;
  FUN_18036d0a0(plVar16,&DAT_1803948e8,local_68,local_88);
  plVar6 = DAT_1803b6a78;
  if (plVar16 == DAT_1803b6a78) {
    if ((-1 < (int)*DAT_1803b6a78) && (*DAT_1803b6a78 = *DAT_1803b6a78 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b6a78 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  plVar16 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar7;
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))();
  }
  return (code *)0x0;
}
