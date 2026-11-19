
/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1801dfdd0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int iVar1;
  int iVar2;
  longlong lVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  int *piVar9;
  code *pcVar10;
  code *pcVar11;
  longlong lVar12;
  undefined8 uVar13;
  undefined4 uVar14;
  longlong lVar15;
  longlong **pplVar16;
  code *pcVar17;
  undefined1 auStack_208 [32];
  longlong *local_1e8;
  longlong **local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  code *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  code *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  ulonglong local_40;
  
  plVar8 = DAT_1803bcfe8;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_208;
  local_148 = (longlong *)*param_3;
  plVar4 = (longlong *)param_3[1];
  pcVar17 = (code *)param_3[10];
  plVar5 = (longlong *)param_3[0xb];
  plVar6 = (longlong *)param_3[0xc];
  local_120 = (longlong *)param_3[2];
  local_110 = (longlong *)param_3[3];
  local_108 = (longlong *)param_3[4];
  local_100 = (longlong *)param_3[5];
  local_f8 = (longlong *)param_3[6];
  local_f0 = (longlong *)param_3[7];
  local_e8 = (longlong *)param_3[8];
  local_128 = (longlong *)param_3[9];
  local_168 = 0;
  local_d8 = (longlong *)0x0;
  local_d0 = (longlong *)0x0;
  local_150 = (longlong *)0x0;
  local_118 = (longlong *)0x0;
  local_e0 = (longlong *)0x0;
  local_138 = DAT_1803bcfe8;
  local_140 = plVar4;
  local_130 = pcVar17;
  local_c8 = plVar5;
  local_c0 = plVar6;
  if (DAT_1803bcfe8 == (longlong *)0x0) {
LAB_1801dfec7:
    DAT_1803bcfe8 = (longlong *)FUN_18036cd10(param_1,DAT_1803bd0f8,DAT_1803ca5e0,0x90);
  }
  else if ((1 < *DAT_1803bcfe8) || (DAT_1803bcfe8[2] != 0)) {
    if ((DAT_1803bcfe8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bcfe8 && (*DAT_1803bcfe8 = *DAT_1803bcfe8 + -1, *plVar8 == 0)))) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    goto LAB_1801dfec7;
  }
  plVar8 = DAT_1803bcfe8;
  local_138 = DAT_1803bcfe8;
  FUN_180001d00(param_1,DAT_1803bcfe8);
  if (pcVar17 == _Py_TrueStruct_exref) {
LAB_1801dffae:
    pcVar10 = pcVar17;
    if (*(int *)pcVar17 + 1 != 0) {
      *(int *)pcVar17 = *(int *)pcVar17 + 1;
    }
LAB_1801dffbe:
    plVar4 = (longlong *)FUN_180383520(pcVar10,DAT_1803bd158);
    if ((-1 < *(int *)pcVar10) &&
       (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
      (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
    }
    if (plVar4 == (longlong *)0x0) {
      lVar3 = *(longlong *)(param_1 + 0x60);
      uVar14 = 0x2e;
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar4 = local_150;
    }
    else {
      local_d8 = plVar4;
      plVar5 = (longlong *)FUN_180383520(plVar5,DAT_1803bd160);
      local_b8 = plVar5;
      if (plVar5 == (longlong *)0x0) {
        lVar3 = *(longlong *)(param_1 + 0x60);
        uVar14 = 0x2f;
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar4 = local_150;
      }
      else {
        local_d0 = plVar5;
        iVar1 = FUN_180001510(plVar6);
        if (iVar1 == -1) {
          lVar3 = *(longlong *)(param_1 + 0x60);
          uVar14 = 0x30;
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar4 = local_150;
        }
        else if (iVar1 == 1) {
          iVar1 = (int)*plVar6 + 1;
          if (iVar1 != 0) {
            *(int *)plVar6 = iVar1;
          }
LAB_1801e0257:
          local_150 = plVar6;
          iVar1 = FUN_180001510(plVar4);
          if (iVar1 == -1) {
            lVar3 = *(longlong *)(param_1 + 0x60);
            uVar14 = 0x34;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar4 = local_150;
          }
          else {
            if (iVar1 == 0) {
              iVar1 = FUN_1803836d0(local_e8,DAT_1803bd180);
              if (iVar1 == -1) {
                lVar3 = *(longlong *)(param_1 + 0x60);
                uVar14 = 0x3d;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar4 = local_150;
                goto LAB_1801e230b;
              }
              lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
              if (iVar1 == 1) {
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar6 = (longlong *)FUN_1800015a0();
                  if (plVar6 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar6;
                  }
LAB_1801e0631:
                  if (lVar12 == 0) goto LAB_1801e0636;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9b40) {
                    DAT_1803a9b40 = iVar1;
                    DAT_1803bce88 =
                         FUN_1803707d0(lVar3,DAT_1803bd198,*(undefined8 *)(DAT_1803bd198 + 0x18));
                  }
                  if (-1 < DAT_1803bce88) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bce88 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e0677;
                    DAT_1803bce88 =
                         FUN_1803707d0(lVar3,DAT_1803bd198,*(undefined8 *)(DAT_1803bd198 + 0x18));
                    if (-1 < DAT_1803bce88) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bce88 * 0x10);
                      goto LAB_1801e0631;
                    }
                  }
LAB_1801e0636:
                  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd198);
                  if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd198);
                    uVar14 = 0x3c;
                    lVar3 = local_168;
                    plVar4 = local_150;
                    goto LAB_1801e230b;
                  }
                }
LAB_1801e0677:
                local_70 = local_148;
                pplVar16 = &local_70;
                local_68 = local_140;
                uVar14 = 0x3c;
                local_60 = local_150;
              }
              else {
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar6 = (longlong *)FUN_1800015a0();
                  if (plVar6 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar6;
                  }
LAB_1801e0764:
                  if (lVar12 == 0) goto LAB_1801e0769;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9b50) {
                    DAT_1803a9b50 = iVar1;
                    DAT_1803bd090 =
                         FUN_1803707d0(lVar3,DAT_1803bd1a0,*(undefined8 *)(DAT_1803bd1a0 + 0x18));
                  }
                  if (-1 < DAT_1803bd090) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bd090 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e07aa;
                    DAT_1803bd090 =
                         FUN_1803707d0(lVar3,DAT_1803bd1a0,*(undefined8 *)(DAT_1803bd1a0 + 0x18));
                    if (-1 < DAT_1803bd090) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bd090 * 0x10);
                      goto LAB_1801e0764;
                    }
                  }
LAB_1801e0769:
                  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd1a0);
                  if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd1a0);
                    uVar14 = 0x3e;
                    lVar3 = local_168;
                    plVar4 = local_150;
                    goto LAB_1801e230b;
                  }
                }
LAB_1801e07aa:
                local_58 = local_148;
                pplVar16 = &local_58;
                local_50 = local_140;
                uVar14 = 0x3e;
                local_48 = local_150;
              }
            }
            else {
              iVar1 = FUN_1803836d0(local_e8,DAT_1803bd180);
              if (iVar1 == -1) {
                lVar3 = *(longlong *)(param_1 + 0x60);
                uVar14 = 0x37;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar4 = local_150;
                goto LAB_1801e230b;
              }
              lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
              if (iVar1 != 1) {
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar6 = (longlong *)FUN_1800015a0();
                  if (plVar6 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar6;
                  }
LAB_1801e04d0:
                  if (lVar12 == 0) goto LAB_1801e04d5;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9ae8) {
                    DAT_1803a9ae8 = iVar1;
                    DAT_1803bd0a0 =
                         FUN_1803707d0(lVar3,DAT_1803bd190,*(undefined8 *)(DAT_1803bd190 + 0x18));
                  }
                  if (-1 < DAT_1803bd0a0) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bd0a0 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e0516;
                    DAT_1803bd0a0 =
                         FUN_1803707d0(lVar3,DAT_1803bd190,*(undefined8 *)(DAT_1803bd190 + 0x18));
                    if (-1 < DAT_1803bd0a0) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bd0a0 * 0x10);
                      goto LAB_1801e04d0;
                    }
                  }
LAB_1801e04d5:
                  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd190);
                  if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd190);
                    uVar14 = 0x38;
                    lVar3 = local_168;
                    plVar4 = local_150;
                    goto LAB_1801e230b;
                  }
                }
LAB_1801e0516:
                local_88 = local_148;
                pplVar16 = &local_88;
                local_80 = local_140;
                uVar14 = 0x38;
                local_78 = local_150;
                goto LAB_1801e03e9;
              }
              if (*(char *)(lVar3 + 10) == '\0') {
                plVar6 = (longlong *)FUN_1800015a0();
                if (plVar6 == (longlong *)0x0) {
                  lVar12 = 0;
                }
                else {
                  lVar12 = *plVar6;
                }
LAB_1801e0382:
                if (lVar12 == 0) goto LAB_1801e0387;
              }
              else {
                iVar1 = *(int *)(lVar3 + 0xc);
                if (iVar1 == 0) {
                  iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                  *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                  *(int *)(lVar3 + 0xc) = iVar1;
                }
                if (iVar1 != DAT_1803a9b4c) {
                  DAT_1803a9b4c = iVar1;
                  DAT_1803bcf80 =
                       FUN_1803707d0(lVar3,DAT_1803bd188,*(undefined8 *)(DAT_1803bd188 + 0x18));
                }
                if (-1 < DAT_1803bcf80) {
                  lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                  lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf80 * 0x10);
                  if (lVar12 != 0) goto LAB_1801e03c8;
                  DAT_1803bcf80 =
                       FUN_1803707d0(lVar3,DAT_1803bd188,*(undefined8 *)(DAT_1803bd188 + 0x18));
                  if (-1 < DAT_1803bcf80) {
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf80 * 0x10);
                    goto LAB_1801e0382;
                  }
                }
LAB_1801e0387:
                plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd188);
                if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd188);
                  uVar14 = 0x36;
                  lVar3 = local_168;
                  plVar4 = local_150;
                  goto LAB_1801e230b;
                }
              }
LAB_1801e03c8:
              pplVar16 = &local_a0;
              uVar14 = 0x36;
              local_a0 = local_148;
              local_90 = local_150;
              local_98 = local_140;
            }
LAB_1801e03e9:
            *(undefined4 *)(plVar8 + 5) = uVar14;
            plVar6 = (longlong *)FUN_180378890(param_1,lVar12,pplVar16);
            if (plVar6 == (longlong *)0x0) {
LAB_1801e03fe:
              lVar3 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar4 = local_150;
            }
            else {
              local_118 = plVar6;
              iVar1 = FUN_180001510(plVar4);
              if (iVar1 == -1) {
                lVar3 = *(longlong *)(param_1 + 0x60);
                uVar14 = 0x41;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar4 = local_150;
                goto LAB_1801e230b;
              }
              lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
              if (iVar1 == 0) {
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar6 = (longlong *)FUN_1800015a0();
                  if (plVar6 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar6;
                  }
LAB_1801e150d:
                  if (lVar12 == 0) goto LAB_1801e1512;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9aec) {
                    DAT_1803a9aec = iVar1;
                    DAT_1803bcf40 =
                         FUN_1803707d0(lVar3,DAT_1803bd280,*(undefined8 *)(DAT_1803bd280 + 0x18));
                  }
                  if (-1 < DAT_1803bcf40) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf40 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e1553;
                    DAT_1803bcf40 =
                         FUN_1803707d0(lVar3,DAT_1803bd280,*(undefined8 *)(DAT_1803bd280 + 0x18));
                    if (-1 < DAT_1803bcf40) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf40 * 0x10);
                      goto LAB_1801e150d;
                    }
                  }
LAB_1801e1512:
                  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd280);
                  if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd280);
                    uVar14 = 0x61;
                    lVar3 = local_168;
                    plVar4 = local_150;
                    goto LAB_1801e230b;
                  }
                }
LAB_1801e1553:
                uVar13 = DAT_1803bd1b0;
                if ((local_148 == (longlong *)0x0) ||
                   (uVar13 = DAT_1803bd1b8, local_140 == (longlong *)0x0)) {
                  plVar4 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
                                      ,uVar13);
                  lVar3 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar4)
                  ;
                  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                    (**(code **)(plVar4[1] + 0x30))(plVar4);
                  }
                  FUN_18000d1d0(param_1,lVar3);
                  uVar14 = 0x61;
                  plVar4 = local_150;
                  goto LAB_1801e230b;
                }
                local_160 = local_148;
                uVar14 = 0x61;
                local_158 = local_140;
                *(undefined4 *)(plVar8 + 5) = 0x61;
                lVar3 = FUN_180377750(param_1,lVar12,&local_160);
                if (lVar3 == 0) goto LAB_1801e03fe;
                plVar6 = (longlong *)FUN_180372160(param_1,0x16);
                *(longlong *)plVar6[3] = lVar3;
                lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
                if (*(char *)(lVar3 + 10) != '\0') {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9b3c) {
                    DAT_1803a9b3c = iVar1;
                    DAT_1803bcf10 =
                         FUN_1803707d0(lVar3,DAT_1803bd288,*(undefined8 *)(DAT_1803bd288 + 0x18));
                  }
                  if (-1 < DAT_1803bcf10) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf10 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e171c;
                    DAT_1803bcf10 =
                         FUN_1803707d0(lVar3,DAT_1803bd288,*(undefined8 *)(DAT_1803bd288 + 0x18));
                    if (-1 < DAT_1803bcf10) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf10 * 0x10);
                      goto LAB_1801e16d6;
                    }
                  }
LAB_1801e16db:
                  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd288);
                  if ((plVar7 != (longlong *)0x0) && (lVar12 = *plVar7, lVar12 != 0))
                  goto LAB_1801e171c;
                  FUN_180373340(param_1,&local_168,DAT_1803bd288);
                  uVar14 = 0x62;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                plVar7 = (longlong *)FUN_1800015a0();
                if (plVar7 == (longlong *)0x0) {
                  lVar12 = 0;
                }
                else {
                  lVar12 = *plVar7;
                }
LAB_1801e16d6:
                if (lVar12 == 0) goto LAB_1801e16db;
LAB_1801e171c:
                plVar7 = local_150;
                if (local_150 == (longlong *)0x0) {
                  plVar4 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
                                      ,DAT_1803bd1e0);
                  lVar3 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar4)
                  ;
                  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                    (**(code **)(plVar4[1] + 0x30))(plVar4);
                  }
                  FUN_18000d1d0(param_1,lVar3);
                  uVar14 = 0x62;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x62;
                *(undefined4 *)(plVar8 + 5) = 0x62;
                lVar3 = FUN_180376770(param_1,lVar12,local_150);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 8) = lVar3;
                lVar3 = FUN_1801de970(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd1c8);
                  uVar14 = 99;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 99;
                local_160 = DAT_1803bd178;
                *(undefined4 *)(plVar8 + 5) = 99;
                local_158 = plVar5;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x10) = lVar3;
                lVar3 = FUN_1801dddc0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd1d0);
                  uVar14 = 100;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 100;
                local_160 = DAT_1803bd178;
                *(undefined4 *)(plVar8 + 5) = 100;
                local_158 = plVar5;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x18) = lVar3;
                lVar3 = FUN_1801deb90(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd1d8);
                  uVar14 = 0x65;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x65;
                local_158 = DAT_1803bd178;
                *(undefined4 *)(plVar8 + 5) = 0x65;
                local_160 = plVar7;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x20) = lVar3;
                lVar3 = FUN_1801df1f0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd1e8);
                  uVar14 = 0x66;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x66;
                *(undefined4 *)(plVar8 + 5) = 0x66;
                lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd290);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x28) = lVar3;
                lVar3 = FUN_1801dedb0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd1f8);
                  uVar14 = 0x67;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x67;
                *(undefined4 *)(plVar8 + 5) = 0x67;
                lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd290);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x30) = lVar3;
                lVar3 = FUN_1801ddfe0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd200);
                  uVar14 = 0x68;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x68;
                *(undefined4 *)(plVar8 + 5) = 0x68;
                lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd290);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x38) = lVar3;
                lVar3 = FUN_1801dea80(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd208);
                  uVar14 = 0x69;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x69;
                *(undefined4 *)(plVar8 + 5) = 0x69;
                lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd290);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x40) = lVar3;
                lVar3 = FUN_1801de860(param_1);
                plVar7 = local_f0;
                plVar5 = DAT_1803bd178;
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd210);
                  uVar14 = 0x6a;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x6a;
                *(undefined4 *)(plVar8 + 5) = 0x6a;
                local_160 = local_f0;
                local_158 = plVar5;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x48) = lVar3;
                lVar3 = FUN_1801de750(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd298);
                  uVar14 = 0x6b;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                iVar1 = FUN_180001510(plVar4);
                uVar14 = 0x6b;
                if (iVar1 == -1) goto LAB_1801e13d9;
                local_160 = plVar7;
                local_158 = DAT_1803bd178;
                if (iVar1 != 0) {
                  local_158 = DAT_1803bd158;
                }
                *(undefined4 *)(plVar8 + 5) = 0x6b;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                plVar5 = local_120;
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x50) = lVar3;
                iVar1 = FUN_180001510(local_120);
                uVar14 = 0x6c;
                if (iVar1 == -1) goto LAB_1801e13d9;
                iVar2 = 0;
                if (iVar1 != 0) {
                  plVar5 = (longlong *)FUN_18037f0c0();
                  if (plVar5 != (longlong *)0x0) {
                    *(undefined4 *)(plVar8 + 5) = 0x6c;
                    plVar7 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803bd2a8);
                    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                    }
                    if (plVar7 != (longlong *)0x0) {
                      iVar2 = FUN_180001510(plVar7);
                      if (iVar2 == -1) {
                        lVar3 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        goto LAB_1801e13e5;
                      }
                      plVar5 = local_120;
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                        plVar5 = local_120;
                      }
                      goto LAB_1801e1c21;
                    }
                  }
                  goto LAB_1801e13d9;
                }
LAB_1801e1c21:
                piVar9 = DAT_1803bd260;
                if (iVar2 == 0) {
                  if (*DAT_1803bd260 + 1 != 0) {
                    *DAT_1803bd260 = *DAT_1803bd260 + 1;
                  }
                }
                else {
                  lVar3 = FUN_1801de530(param_1);
                  if (lVar3 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd2b0);
                    lVar3 = local_168;
                    goto LAB_1801e13e5;
                  }
                  *(undefined4 *)(plVar8 + 5) = 0x6c;
                  piVar9 = (int *)FUN_180376210(param_1,lVar3);
                  if (piVar9 == (int *)0x0) goto LAB_1801e13d9;
                }
                *(int **)(plVar6[3] + 0x58) = piVar9;
                lVar3 = FUN_1801de640(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd218);
                  uVar14 = 0x6d;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                local_160 = local_140;
                uVar14 = 0x6d;
                local_158 = local_108;
                *(undefined4 *)(plVar8 + 5) = 0x6d;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x60) = lVar3;
                lVar3 = FUN_1801deec0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd220);
                  uVar14 = 0x6e;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x6e;
                *(undefined4 *)(plVar8 + 5) = 0x6e;
                lVar3 = FUN_180376770(param_1,lVar3,plVar5);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x68) = lVar3;
                lVar3 = FUN_1801deca0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd228);
                  uVar14 = 0x6f;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                iVar1 = FUN_180001510(plVar4);
                uVar14 = 0x6f;
                if (iVar1 == -1) goto LAB_1801e13d9;
                local_158 = DAT_1803bd178;
                if (iVar1 != 0) {
                  local_158 = DAT_1803bd158;
                }
                *(undefined4 *)(plVar8 + 5) = 0x6f;
                local_160 = plVar5;
                lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x70) = lVar3;
                lVar3 = FUN_1801ddcb0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd230);
                  uVar14 = 0x70;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x70;
                *(undefined4 *)(plVar8 + 5) = 0x70;
                lVar3 = FUN_180376770(param_1,lVar3,local_110);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x78) = lVar3;
                lVar3 = FUN_1801df0e0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd238);
                  uVar14 = 0x71;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x71;
                *(undefined4 *)(plVar8 + 5) = 0x71;
                lVar3 = FUN_180376770(param_1,lVar3,local_100);
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x80) = lVar3;
                lVar3 = FUN_1801defd0(param_1);
                if (lVar3 == 0) {
                  FUN_180373340(param_1,&local_168,DAT_1803bd240);
                  uVar14 = 0x72;
                  lVar3 = local_168;
                  goto LAB_1801e13e5;
                }
                uVar14 = 0x72;
                *(undefined4 *)(plVar8 + 5) = 0x72;
                lVar3 = FUN_180376770(param_1,lVar3,local_f8);
                plVar5 = local_128;
                if (lVar3 == 0) goto LAB_1801e13d9;
                *(longlong *)(plVar6[3] + 0x88) = lVar3;
                iVar1 = FUN_180001510(local_128);
                piVar9 = DAT_1803bd260;
                if (iVar1 == -1) {
                  uVar14 = 0x74;
                  goto LAB_1801e13d9;
                }
                if (iVar1 != 0) {
                  lVar3 = FUN_1801de310(param_1);
                  if (lVar3 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd248);
                    uVar14 = 0x73;
                    lVar3 = local_168;
                    goto LAB_1801e13e5;
                  }
                  plVar8 = (longlong *)FUN_180024e50(param_1,plVar5,DAT_1803bd250);
                  if (plVar8 == (longlong *)0x0) {
                    uVar14 = 0x73;
                  }
                  else {
                    plVar5 = (longlong *)FUN_180024e50(param_1,plVar5,DAT_1803bd258);
                    if (plVar5 == (longlong *)0x0) {
                      lVar3 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                      uVar14 = 0x73;
                      goto LAB_1801e13e5;
                    }
                    uVar14 = 0x73;
                    *(undefined4 *)(local_138 + 5) = 0x73;
                    local_160 = plVar8;
                    local_158 = plVar5;
                    piVar9 = (int *)FUN_180377750(param_1,lVar3,&local_160);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                    }
                    plVar5 = local_128;
                    if (piVar9 != (int *)0x0) goto LAB_1801e1ff8;
                  }
                  goto LAB_1801e13d9;
                }
                if (*DAT_1803bd260 + 1 != 0) {
                  *DAT_1803bd260 = *DAT_1803bd260 + 1;
                }
LAB_1801e1ff8:
                *(int **)(plVar6[3] + 0x90) = piVar9;
                iVar1 = FUN_180001510(plVar5);
                piVar9 = DAT_1803bd260;
                if (iVar1 == -1) {
                  uVar14 = 0x76;
                  goto LAB_1801e13d9;
                }
                if (iVar1 != 0) {
                  lVar3 = FUN_1801dded0(param_1);
                  if (lVar3 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd268);
                    uVar14 = 0x76;
                    lVar3 = local_168;
                    goto LAB_1801e13e5;
                  }
                  plVar5 = (longlong *)FUN_180024e50(param_1,plVar5,DAT_1803bd250);
                  uVar14 = 0x76;
                  if (plVar5 != (longlong *)0x0) {
                    *(undefined4 *)(local_138 + 5) = 0x76;
                    piVar9 = (int *)FUN_180376770(param_1,lVar3,plVar5);
                    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                    }
                    plVar5 = local_128;
                    if (piVar9 != (int *)0x0) goto LAB_1801e20c5;
                  }
                  goto LAB_1801e13d9;
                }
                if (*DAT_1803bd260 + 1 != 0) {
                  *DAT_1803bd260 = *DAT_1803bd260 + 1;
                }
LAB_1801e20c5:
                *(int **)(plVar6[3] + 0x98) = piVar9;
                iVar1 = FUN_180001510(plVar5);
                piVar9 = DAT_1803bd260;
                if (iVar1 == -1) {
                  uVar14 = 0x7a;
                  goto LAB_1801e13d9;
                }
                if (iVar1 != 0) {
                  lVar3 = FUN_1801de420(param_1);
                  if (lVar3 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd270);
                    uVar14 = 0x77;
                    lVar3 = local_168;
                    goto LAB_1801e13e5;
                  }
                  plVar8 = (longlong *)FUN_180024e50(param_1,plVar5,DAT_1803bd258);
                  if (plVar8 == (longlong *)0x0) {
                    uVar14 = 0x78;
                  }
                  else {
                    plVar5 = (longlong *)FUN_180024e50(param_1,plVar5,DAT_1803bd278);
                    if (plVar5 == (longlong *)0x0) {
                      lVar3 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                      uVar14 = 0x78;
                      goto LAB_1801e13e5;
                    }
                    uVar14 = 0x77;
                    *(undefined4 *)(local_138 + 5) = 0x77;
                    local_160 = plVar8;
                    local_158 = plVar5;
                    piVar9 = (int *)FUN_180377750(param_1,lVar3,&local_160);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                    }
                    if (piVar9 != (int *)0x0) goto LAB_1801e220d;
                  }
                  goto LAB_1801e13d9;
                }
                if (*DAT_1803bd260 + 1 != 0) {
                  *DAT_1803bd260 = *DAT_1803bd260 + 1;
                }
LAB_1801e220d:
                *(int **)(plVar6[3] + 0xa0) = piVar9;
                iVar1 = (int)*local_118 + 1;
                if (iVar1 != 0) {
                  *(int *)local_118 = iVar1;
                }
                *(longlong **)(plVar6[3] + 0xa8) = local_118;
LAB_1801e2234:
                plVar5 = local_118;
                uVar13 = DAT_1803bd2b8;
                local_e0 = plVar6;
                plVar8 = (longlong *)FUN_180001930();
                if (plVar8 != (longlong *)0x0) {
                  local_160 = (longlong *)FUN_18036d750(plVar8);
                  local_1e0 = &local_160;
                  local_1d0 = (longlong *)0x48;
                  local_1d8 = (longlong *)0x1;
                  local_1e8 = (longlong *)DAT_1803bcf98;
                  plVar7 = (longlong *)
                           FUN_180367890(FUN_1801e28a0,DAT_1803ca5e0,DAT_1803bd2c0,DAT_1803bd2c8);
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  lVar3 = PyUnicode_Join(uVar13,plVar7);
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  if (lVar3 != 0) {
                    FUN_180001d80(param_1);
                    if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                      (**(code **)(plVar4[1] + 0x30))(plVar4);
                    }
                    if ((-1 < (int)*local_b8) && (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
                      (**(code **)(local_d0[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_150) && (*local_150 = *local_150 + -1, *local_150 == 0)) {
                      (**(code **)(local_150[1] + 0x30))();
                    }
                    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                    }
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                    if ((-1 < (int)*local_148) && (*local_148 = *local_148 + -1, *local_148 == 0)) {
                      (**(code **)(local_148[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_140) && (*local_140 = *local_140 + -1, *local_140 == 0)) {
                      (**(code **)(local_140[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_120) && (*local_120 = *local_120 + -1, *local_120 == 0)) {
                      (**(code **)(local_120[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_110) && (*local_110 = *local_110 + -1, *local_110 == 0)) {
                      (**(code **)(local_110[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_108) && (*local_108 = *local_108 + -1, *local_108 == 0)) {
                      (**(code **)(local_108[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_100) && (*local_100 = *local_100 + -1, *local_100 == 0)) {
                      (**(code **)(local_100[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_f8) && (*local_f8 = *local_f8 + -1, *local_f8 == 0)) {
                      (**(code **)(local_f8[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_f0) && (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
                      (**(code **)(local_f0[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_e8) && (*local_e8 = *local_e8 + -1, *local_e8 == 0)) {
                      (**(code **)(local_e8[1] + 0x30))();
                    }
                    if ((-1 < (int)*local_128) && (*local_128 = *local_128 + -1, *local_128 == 0)) {
                      (**(code **)(local_128[1] + 0x30))();
                    }
                    if ((-1 < *(int *)local_130) &&
                       (*(longlong *)local_130 = *(longlong *)local_130 + -1,
                       *(longlong *)local_130 == 0)) {
                      (**(code **)(*(longlong *)(local_130 + 8) + 0x30))();
                    }
                    if ((-1 < (int)*local_c8) && (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
                      (**(code **)(local_c8[1] + 0x30))();
                    }
                    if ((int)*local_c0 < 0) {
                      return lVar3;
                    }
                    *local_c0 = *local_c0 + -1;
                    if (*local_c0 != 0) {
                      return lVar3;
                    }
                    (**(code **)(local_c0[1] + 0x30))();
                    return lVar3;
                  }
                }
                lVar3 = *(longlong *)(param_1 + 0x60);
                uVar14 = 0x80;
                *(undefined8 *)(param_1 + 0x60) = 0;
                pcVar17 = local_130;
                plVar4 = local_150;
              }
              else {
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar6 = (longlong *)FUN_1800015a0();
                  if (plVar6 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar6;
                  }
LAB_1801e08d7:
                  if (lVar12 == 0) goto LAB_1801e08dc;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9b58) {
                    DAT_1803a9b58 = iVar1;
                    DAT_1803bced0 =
                         FUN_1803707d0(lVar3,DAT_1803bd1a8,*(undefined8 *)(DAT_1803bd1a8 + 0x18));
                  }
                  if (-1 < DAT_1803bced0) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bced0 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e091d;
                    DAT_1803bced0 =
                         FUN_1803707d0(lVar3,DAT_1803bd1a8,*(undefined8 *)(DAT_1803bd1a8 + 0x18));
                    if (-1 < DAT_1803bced0) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bced0 * 0x10);
                      goto LAB_1801e08d7;
                    }
                  }
LAB_1801e08dc:
                  plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd1a8);
                  if ((plVar6 == (longlong *)0x0) || (lVar12 = *plVar6, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd1a8);
                    uVar14 = 0x43;
                    lVar3 = local_168;
                    plVar4 = local_150;
                    goto LAB_1801e230b;
                  }
                }
LAB_1801e091d:
                uVar13 = DAT_1803bd1b0;
                if ((local_148 == (longlong *)0x0) ||
                   (uVar13 = DAT_1803bd1b8, local_140 == (longlong *)0x0)) {
                  plVar4 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
                                      ,uVar13);
                  lVar3 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar4)
                  ;
                  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                    (**(code **)(plVar4[1] + 0x30))(plVar4);
                  }
                  FUN_18000d1d0(param_1,lVar3);
                  uVar14 = 0x43;
                  plVar4 = local_150;
                  goto LAB_1801e230b;
                }
                local_160 = local_148;
                uVar14 = 0x43;
                local_158 = local_140;
                *(undefined4 *)(plVar8 + 5) = 0x43;
                lVar3 = FUN_180377750(param_1,lVar12,&local_160);
                if (lVar3 == 0) goto LAB_1801e03fe;
                plVar6 = (longlong *)FUN_180372160(param_1,0x14);
                *(longlong *)plVar6[3] = lVar3;
                lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
                if (*(char *)(lVar3 + 10) == '\0') {
                  plVar7 = (longlong *)FUN_1800015a0();
                  if (plVar7 == (longlong *)0x0) {
                    lVar12 = 0;
                  }
                  else {
                    lVar12 = *plVar7;
                  }
LAB_1801e0aa0:
                  if (lVar12 == 0) goto LAB_1801e0aa5;
                }
                else {
                  iVar1 = *(int *)(lVar3 + 0xc);
                  if (iVar1 == 0) {
                    iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
                    *(int *)(lVar3 + 0xc) = iVar1;
                  }
                  if (iVar1 != DAT_1803a9b08) {
                    DAT_1803a9b08 = iVar1;
                    DAT_1803bcf08 =
                         FUN_1803707d0(lVar3,DAT_1803bd1c0,*(undefined8 *)(DAT_1803bd1c0 + 0x18));
                  }
                  if (-1 < DAT_1803bcf08) {
                    lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
                    lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf08 * 0x10);
                    if (lVar12 != 0) goto LAB_1801e0ae6;
                    DAT_1803bcf08 =
                         FUN_1803707d0(lVar3,DAT_1803bd1c0,*(undefined8 *)(DAT_1803bd1c0 + 0x18));
                    if (-1 < DAT_1803bcf08) {
                      lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcf08 * 0x10);
                      goto LAB_1801e0aa0;
                    }
                  }
LAB_1801e0aa5:
                  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd1c0);
                  if ((plVar7 == (longlong *)0x0) || (lVar12 = *plVar7, lVar12 == 0)) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd1c0);
                    uVar14 = 0x44;
                    lVar3 = local_168;
                    goto LAB_1801e13e5;
                  }
                }
LAB_1801e0ae6:
                uVar14 = 0x44;
                *(undefined4 *)(plVar8 + 5) = 0x44;
                lVar3 = FUN_180376770(param_1,lVar12,local_148);
                if (lVar3 == 0) {
LAB_1801e13d9:
                  lVar3 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  *(longlong *)(plVar6[3] + 8) = lVar3;
                  lVar3 = FUN_1801de970(param_1);
                  if (lVar3 == 0) {
                    FUN_180373340(param_1,&local_168,DAT_1803bd1c8);
                    uVar14 = 0x45;
                    lVar3 = local_168;
                  }
                  else {
                    uVar14 = 0x45;
                    local_160 = DAT_1803bd158;
                    *(undefined4 *)(plVar8 + 5) = 0x45;
                    local_158 = plVar5;
                    lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                    if (lVar3 == 0) goto LAB_1801e13d9;
                    *(longlong *)(plVar6[3] + 0x10) = lVar3;
                    lVar3 = FUN_1801dddc0(param_1);
                    if (lVar3 == 0) {
                      FUN_180373340(param_1,&local_168,DAT_1803bd1d0);
                      uVar14 = 0x46;
                      lVar3 = local_168;
                    }
                    else {
                      uVar14 = 0x46;
                      local_160 = DAT_1803bd158;
                      *(undefined4 *)(plVar8 + 5) = 0x46;
                      local_158 = plVar5;
                      lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                      if (lVar3 == 0) goto LAB_1801e13d9;
                      *(longlong *)(plVar6[3] + 0x18) = lVar3;
                      lVar3 = FUN_1801deb90(param_1);
                      plVar5 = DAT_1803bd158;
                      if (lVar3 == 0) {
                        FUN_180373340(param_1,&local_168,DAT_1803bd1d8);
                        uVar14 = 0x47;
                        lVar3 = local_168;
                      }
                      else {
                        if (local_150 != (longlong *)0x0) {
                          local_160 = local_150;
                          uVar14 = 0x47;
                          *(undefined4 *)(plVar8 + 5) = 0x47;
                          local_158 = plVar5;
                          lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                          if (lVar3 != 0) {
                            *(longlong *)(plVar6[3] + 0x20) = lVar3;
                            lVar3 = FUN_1801df1f0(param_1);
                            if (lVar3 == 0) {
                              FUN_180373340(param_1,&local_168,DAT_1803bd1e8);
                              uVar14 = 0x48;
                              lVar3 = local_168;
                              goto LAB_1801e13e5;
                            }
                            *(undefined4 *)(plVar8 + 5) = 0x48;
                            lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd1f0);
                            if (lVar3 == 0) {
                              uVar14 = 0x48;
                            }
                            else {
                              *(longlong *)(plVar6[3] + 0x28) = lVar3;
                              lVar3 = FUN_1801dedb0(param_1);
                              if (lVar3 == 0) {
                                FUN_180373340(param_1,&local_168,DAT_1803bd1f8);
                                uVar14 = 0x49;
                                lVar3 = local_168;
                                goto LAB_1801e13e5;
                              }
                              uVar14 = 0x49;
                              *(undefined4 *)(plVar8 + 5) = 0x49;
                              lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd1f0);
                              if (lVar3 != 0) {
                                *(longlong *)(plVar6[3] + 0x30) = lVar3;
                                lVar3 = FUN_1801ddfe0(param_1);
                                if (lVar3 == 0) {
                                  FUN_180373340(param_1,&local_168,DAT_1803bd200);
                                  uVar14 = 0x4a;
                                  lVar3 = local_168;
                                  goto LAB_1801e13e5;
                                }
                                uVar14 = 0x4a;
                                *(undefined4 *)(plVar8 + 5) = 0x4a;
                                lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd1f0);
                                if (lVar3 != 0) {
                                  *(longlong *)(plVar6[3] + 0x38) = lVar3;
                                  lVar3 = FUN_1801dea80(param_1);
                                  if (lVar3 == 0) {
                                    FUN_180373340(param_1,&local_168,DAT_1803bd208);
                                    uVar14 = 0x4b;
                                    lVar3 = local_168;
                                    goto LAB_1801e13e5;
                                  }
                                  uVar14 = 0x4b;
                                  *(undefined4 *)(plVar8 + 5) = 0x4b;
                                  lVar3 = FUN_1803770d0(param_1,lVar3,DAT_1803bd1f0);
                                  if (lVar3 != 0) {
                                    *(longlong *)(plVar6[3] + 0x40) = lVar3;
                                    lVar3 = FUN_1801de860(param_1);
                                    plVar5 = DAT_1803bd158;
                                    if (lVar3 == 0) {
                                      FUN_180373340(param_1,&local_168,DAT_1803bd210);
                                      uVar14 = 0x4c;
                                      lVar3 = local_168;
                                      goto LAB_1801e13e5;
                                    }
                                    uVar14 = 0x4c;
                                    local_160 = local_f0;
                                    *(undefined4 *)(plVar8 + 5) = 0x4c;
                                    local_158 = plVar5;
                                    lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                                    if (lVar3 != 0) {
                                      *(longlong *)(plVar6[3] + 0x48) = lVar3;
                                      lVar3 = FUN_1801de640(param_1);
                                      if (lVar3 == 0) {
                                        FUN_180373340(param_1,&local_168,DAT_1803bd218);
                                        uVar14 = 0x4d;
                                        lVar3 = local_168;
                                        goto LAB_1801e13e5;
                                      }
                                      local_160 = local_140;
                                      uVar14 = 0x4d;
                                      local_158 = local_108;
                                      *(undefined4 *)(plVar8 + 5) = 0x4d;
                                      lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                                      if (lVar3 != 0) {
                                        *(longlong *)(plVar6[3] + 0x50) = lVar3;
                                        lVar3 = FUN_1801deec0(param_1);
                                        plVar5 = local_120;
                                        if (lVar3 == 0) {
                                          FUN_180373340(param_1,&local_168,DAT_1803bd220);
                                          uVar14 = 0x4e;
                                          lVar3 = local_168;
                                          goto LAB_1801e13e5;
                                        }
                                        uVar14 = 0x4e;
                                        *(undefined4 *)(plVar8 + 5) = 0x4e;
                                        lVar3 = FUN_180376770(param_1,lVar3,local_120);
                                        if (lVar3 != 0) {
                                          *(longlong *)(plVar6[3] + 0x58) = lVar3;
                                          lVar3 = FUN_1801deca0(param_1);
                                          if (lVar3 == 0) {
                                            FUN_180373340(param_1,&local_168,DAT_1803bd228);
                                            uVar14 = 0x4f;
                                            lVar3 = local_168;
                                            goto LAB_1801e13e5;
                                          }
                                          uVar14 = 0x4f;
                                          local_158 = DAT_1803bd158;
                                          *(undefined4 *)(plVar8 + 5) = 0x4f;
                                          local_160 = plVar5;
                                          lVar3 = FUN_180377750(param_1,lVar3,&local_160);
                                          if (lVar3 != 0) {
                                            *(longlong *)(plVar6[3] + 0x60) = lVar3;
                                            lVar3 = FUN_1801ddcb0(param_1);
                                            if (lVar3 == 0) {
                                              FUN_180373340(param_1,&local_168,DAT_1803bd230);
                                              uVar14 = 0x50;
                                              lVar3 = local_168;
                                              goto LAB_1801e13e5;
                                            }
                                            uVar14 = 0x50;
                                            *(undefined4 *)(plVar8 + 5) = 0x50;
                                            lVar3 = FUN_180376770(param_1,lVar3,local_110);
                                            if (lVar3 != 0) {
                                              *(longlong *)(plVar6[3] + 0x68) = lVar3;
                                              lVar3 = FUN_1801df0e0(param_1);
                                              if (lVar3 == 0) {
                                                FUN_180373340(param_1,&local_168,DAT_1803bd238);
                                                uVar14 = 0x51;
                                                lVar3 = local_168;
                                                goto LAB_1801e13e5;
                                              }
                                              uVar14 = 0x51;
                                              *(undefined4 *)(plVar8 + 5) = 0x51;
                                              lVar3 = FUN_180376770(param_1,lVar3,local_100);
                                              if (lVar3 != 0) {
                                                *(longlong *)(plVar6[3] + 0x70) = lVar3;
                                                lVar3 = FUN_1801defd0(param_1);
                                                if (lVar3 == 0) {
                                                  FUN_180373340(param_1,&local_168,DAT_1803bd240);
                                                  uVar14 = 0x52;
                                                  lVar3 = local_168;
                                                  goto LAB_1801e13e5;
                                                }
                                                uVar14 = 0x52;
                                                *(undefined4 *)(plVar8 + 5) = 0x52;
                                                lVar3 = FUN_180376770(param_1,lVar3,local_f8);
                                                plVar5 = local_128;
                                                if (lVar3 != 0) {
                                                  *(longlong *)(plVar6[3] + 0x78) = lVar3;
                                                  iVar1 = FUN_180001510(local_128);
                                                  piVar9 = DAT_1803bd260;
                                                  if (iVar1 == -1) {
                                                    uVar14 = 0x54;
                                                  }
                                                  else if (iVar1 == 0) {
                                                    if (*DAT_1803bd260 + 1 != 0) {
                                                      *DAT_1803bd260 = *DAT_1803bd260 + 1;
                                                    }
LAB_1801e11de:
                                                    *(int **)(plVar6[3] + 0x80) = piVar9;
                                                    iVar1 = FUN_180001510(plVar5);
                                                    piVar9 = DAT_1803bd260;
                                                    if (iVar1 == -1) {
                                                      uVar14 = 0x56;
                                                    }
                                                    else if (iVar1 == 0) {
                                                      if (*DAT_1803bd260 + 1 != 0) {
                                                        *DAT_1803bd260 = *DAT_1803bd260 + 1;
                                                      }
LAB_1801e12ab:
                                                      *(int **)(plVar6[3] + 0x88) = piVar9;
                                                      iVar1 = FUN_180001510(plVar5);
                                                      piVar9 = DAT_1803bd260;
                                                      if (iVar1 == -1) {
                                                        uVar14 = 0x5a;
                                                      }
                                                      else {
                                                        if (iVar1 == 0) {
                                                          if (*DAT_1803bd260 + 1 != 0) {
                                                            *DAT_1803bd260 = *DAT_1803bd260 + 1;
                                                          }
LAB_1801e1420:
                                                          *(int **)(plVar6[3] + 0x90) = piVar9;
                                                          iVar1 = (int)*local_118 + 1;
                                                          if (iVar1 != 0) {
                                                            *(int *)local_118 = iVar1;
                                                          }
                                                          *(longlong **)(plVar6[3] + 0x98) =
                                                               local_118;
                                                          goto LAB_1801e2234;
                                                        }
                                                        lVar3 = FUN_1801de420(param_1);
                                                        if (lVar3 == 0) {
                                                          FUN_180373340(param_1,&local_168,
                                                                        DAT_1803bd270);
                                                          uVar14 = 0x57;
                                                          lVar3 = local_168;
                                                          goto LAB_1801e13e5;
                                                        }
                                                        plVar8 = (longlong *)
                                                                 FUN_180024e50(param_1,plVar5,
                                                                               DAT_1803bd258);
                                                        if (plVar8 == (longlong *)0x0) {
                                                          uVar14 = 0x58;
                                                        }
                                                        else {
                                                          plVar5 = (longlong *)
                                                                   FUN_180024e50(param_1,plVar5,
                                                                                 DAT_1803bd278);
                                                          if (plVar5 == (longlong *)0x0) {
                                                            lVar3 = *(longlong *)(param_1 + 0x60);
                                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              (**(code **)(plVar8[1] + 0x30))
                                                                        (plVar8);
                                                            }
                                                            uVar14 = 0x58;
                                                            goto LAB_1801e13e5;
                                                          }
                                                          uVar14 = 0x57;
                                                          *(undefined4 *)(local_138 + 5) = 0x57;
                                                          local_160 = plVar8;
                                                          local_158 = plVar5;
                                                          piVar9 = (int *)FUN_180377750(param_1,
                                                  lVar3,&local_160);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                  }
                                                  if (piVar9 != (int *)0x0) goto LAB_1801e1420;
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    lVar3 = FUN_1801dded0(param_1);
                                                    if (lVar3 == 0) {
                                                      FUN_180373340(param_1,&local_168,DAT_1803bd268
                                                                   );
                                                      uVar14 = 0x56;
                                                      lVar3 = local_168;
                                                      goto LAB_1801e13e5;
                                                    }
                                                    plVar5 = (longlong *)
                                                             FUN_180024e50(param_1,plVar5,
                                                                           DAT_1803bd250);
                                                    uVar14 = 0x56;
                                                    if (plVar5 != (longlong *)0x0) {
                                                      *(undefined4 *)(local_138 + 5) = 0x56;
                                                      piVar9 = (int *)FUN_180376770(param_1,lVar3,
                                                                                    plVar5);
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                      }
                                                      plVar5 = local_128;
                                                      if (piVar9 != (int *)0x0) goto LAB_1801e12ab;
                                                    }
                                                  }
                                                  }
                                                  else {
                                                    lVar3 = FUN_1801de310(param_1);
                                                    if (lVar3 == 0) {
                                                      FUN_180373340(param_1,&local_168,DAT_1803bd248
                                                                   );
                                                      uVar14 = 0x53;
                                                      lVar3 = local_168;
                                                      goto LAB_1801e13e5;
                                                    }
                                                    plVar8 = (longlong *)
                                                             FUN_180024e50(param_1,plVar5,
                                                                           DAT_1803bd250);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      uVar14 = 0x53;
                                                    }
                                                    else {
                                                      plVar5 = (longlong *)
                                                               FUN_180024e50(param_1,plVar5,
                                                                             DAT_1803bd258);
                                                      if (plVar5 == (longlong *)0x0) {
                                                        lVar3 = *(longlong *)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                        }
                                                        uVar14 = 0x53;
                                                        goto LAB_1801e13e5;
                                                      }
                                                      uVar14 = 0x53;
                                                      *(undefined4 *)(local_138 + 5) = 0x53;
                                                      local_160 = plVar8;
                                                      local_158 = plVar5;
                                                      piVar9 = (int *)FUN_180377750(param_1,lVar3,
                                                                                    &local_160);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                      }
                                                      plVar5 = local_128;
                                                      if (piVar9 != (int *)0x0) goto LAB_1801e11de;
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                          goto LAB_1801e13d9;
                        }
                        plVar4 = (longlong *)
                                 PyUnicode_FromFormat
                                           ("cannot access local variable \'%U\' where it is not associated with a value"
                                            ,DAT_1803bd1e0);
                        lVar3 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,
                                              plVar4);
                        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                          (**(code **)(plVar4[1] + 0x30))(plVar4);
                        }
                        FUN_18000d1d0(param_1,lVar3);
                        uVar14 = 0x47;
                      }
                    }
                  }
                }
LAB_1801e13e5:
                pcVar17 = local_130;
                plVar4 = local_150;
                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                  pcVar17 = local_130;
                  plVar4 = local_150;
                }
              }
            }
          }
        }
        else {
          lVar3 = *(longlong *)(DAT_1803ca5d8 + 0x20);
          if (*(char *)(lVar3 + 10) == '\0') {
            plVar5 = (longlong *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bd168);
            if (plVar5 == (longlong *)0x0) {
              lVar12 = 0;
            }
            else {
              lVar12 = *plVar5;
            }
LAB_1801e013b:
            if (lVar12 == 0) goto LAB_1801e0140;
          }
          else {
            iVar1 = *(int *)(lVar3 + 0xc);
            if (iVar1 == 0) {
              iVar1 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar1 + 1;
              *(int *)(lVar3 + 0xc) = iVar1;
            }
            if (iVar1 != DAT_1803a9b48) {
              DAT_1803a9b48 = iVar1;
              DAT_1803bcff8 =
                   FUN_1803707d0(lVar3,DAT_1803bd168,*(undefined8 *)(DAT_1803bd168 + 0x18));
            }
            if (-1 < DAT_1803bcff8) {
              lVar15 = (1L << (*(byte *)(lVar3 + 9) & 0x3f)) + lVar3;
              lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcff8 * 0x10);
              if (lVar12 != 0) goto LAB_1801e0181;
              DAT_1803bcff8 =
                   FUN_1803707d0(lVar3,DAT_1803bd168,*(undefined8 *)(DAT_1803bd168 + 0x18));
              if (-1 < DAT_1803bcff8) {
                lVar12 = *(longlong *)(lVar15 + 0x28 + DAT_1803bcff8 * 0x10);
                goto LAB_1801e013b;
              }
            }
LAB_1801e0140:
            plVar5 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd168);
            if ((plVar5 == (longlong *)0x0) || (lVar12 = *plVar5, lVar12 == 0)) {
              FUN_180373340(param_1,&local_168,DAT_1803bd168);
              uVar14 = 0x30;
              lVar3 = local_168;
              plVar4 = local_150;
              goto LAB_1801e230b;
            }
          }
LAB_1801e0181:
          plVar5 = (longlong *)FUN_180024e50(param_1,local_148,DAT_1803bd170);
          if (plVar5 == (longlong *)0x0) {
            lVar3 = *(longlong *)(param_1 + 0x60);
            uVar14 = 0x31;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar4 = local_150;
          }
          else {
            iVar1 = FUN_180001510(plVar4);
            if (iVar1 == -1) {
              lVar3 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                (**(code **)(plVar5[1] + 0x30))(plVar5);
              }
              uVar14 = 0x31;
              plVar4 = local_150;
            }
            else {
              local_a8 = DAT_1803bd178;
              if (iVar1 != 0) {
                local_a8 = DAT_1803bd158;
              }
              uVar14 = 0x30;
              *(undefined4 *)(plVar8 + 5) = 0x30;
              local_b0 = plVar5;
              plVar6 = (longlong *)FUN_180377750(param_1,lVar12,&local_b0);
              if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                (**(code **)(plVar5[1] + 0x30))(plVar5);
              }
              plVar5 = local_b8;
              if (plVar6 != (longlong *)0x0) goto LAB_1801e0257;
              lVar3 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar4 = local_150;
            }
          }
        }
      }
    }
  }
  else if ((pcVar17 == _Py_FalseStruct_exref) || (pcVar17 == _Py_NoneStruct_exref)) {
LAB_1801dff82:
    pcVar10 = (code *)FUN_180024e50(param_1,plVar4,DAT_1803bd150);
    if (pcVar10 != (code *)0x0) goto LAB_1801dffbe;
    lVar3 = *(longlong *)(param_1 + 0x60);
    uVar14 = 0x2e;
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar4 = (longlong *)0x0;
  }
  else {
    lVar3 = *(longlong *)(pcVar17 + 8);
    if ((*(longlong *)(lVar3 + 0x60) == 0) ||
       (pcVar10 = *(code **)(*(longlong *)(lVar3 + 0x60) + 0x48), pcVar10 == (code *)0x0)) {
      if ((*(undefined8 **)(lVar3 + 0x70) == (undefined8 *)0x0) ||
         (pcVar10 = (code *)**(undefined8 **)(lVar3 + 0x70), pcVar10 == (code *)0x0)) {
        if ((*(undefined8 **)(lVar3 + 0x68) == (undefined8 *)0x0) ||
           (pcVar10 = (code *)**(undefined8 **)(lVar3 + 0x68), pcVar10 == (code *)0x0))
        goto LAB_1801dffae;
        lVar3 = (*pcVar10)(pcVar17);
      }
      else {
        lVar3 = (*pcVar10)(pcVar17);
      }
    }
    else {
      iVar1 = (*pcVar10)(pcVar17);
      lVar3 = (longlong)iVar1;
    }
    if (0 < lVar3) goto LAB_1801dffae;
    if (lVar3 == 0) goto LAB_1801dff82;
    lVar3 = *(longlong *)(param_1 + 0x60);
    uVar14 = 0x2e;
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar4 = (longlong *)0x0;
  }
LAB_1801e230b:
  piVar9 = *(int **)(lVar3 + 0x28);
  if (piVar9 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(local_138,uVar14);
    plVar5 = *(longlong **)(lVar3 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1801e236a:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar9 + 6) == local_138) goto LAB_1801e23b3;
    pcVar10 = (code *)FUN_18038e2a0(local_138,uVar14);
    *(int **)(pcVar10 + 0x10) = piVar9;
    if (*piVar9 + 1 != 0) {
      *piVar9 = *piVar9 + 1;
    }
    plVar5 = *(longlong **)(lVar3 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1801e236a;
    pcVar11 = (code *)0x0;
  }
  *(code **)(lVar3 + 0x28) = pcVar11;
  if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)
     ) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1801e23b3:
  plVar8 = local_118;
  plVar6 = local_128;
  local_170 = local_e0;
  local_178 = local_118;
  local_188 = local_d0;
  local_190 = local_d8;
  local_198 = local_c0;
  local_1a0 = local_c8;
  local_1b0 = local_128;
  local_1b8 = local_e8;
  local_1c0 = local_f0;
  local_1c8 = local_f8;
  local_1d0 = local_100;
  local_1d8 = local_108;
  local_1e0 = (longlong **)local_110;
  local_1e8 = local_120;
  local_1a8 = pcVar17;
  local_180 = plVar4;
  FUN_18036d0a0(local_138,"oooooooooooooooooo",local_148,local_140);
  plVar5 = DAT_1803bcfe8;
  if (local_138 == DAT_1803bcfe8) {
    if ((-1 < (int)*DAT_1803bcfe8) && (*DAT_1803bcfe8 = *DAT_1803bcfe8 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    DAT_1803bcfe8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_d8 != (longlong *)0x0) && (-1 < (int)*local_d8)) &&
     (*local_d8 = *local_d8 + -1, *local_d8 == 0)) {
    (**(code **)(local_d8[1] + 0x30))();
  }
  if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
     (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
  }
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
     (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    (**(code **)(local_e0[1] + 0x30))();
  }
  if ((-1 < (int)*local_148) && (*local_148 = *local_148 + -1, *local_148 == 0)) {
    (**(code **)(local_148[1] + 0x30))();
  }
  if ((-1 < (int)*local_140) && (*local_140 = *local_140 + -1, *local_140 == 0)) {
    (**(code **)(local_140[1] + 0x30))();
  }
  if ((-1 < (int)*local_120) && (*local_120 = *local_120 + -1, *local_120 == 0)) {
    (**(code **)(local_120[1] + 0x30))();
  }
  if ((-1 < (int)*local_110) && (*local_110 = *local_110 + -1, *local_110 == 0)) {
    (**(code **)(local_110[1] + 0x30))();
  }
  if ((-1 < (int)*local_108) && (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  if ((-1 < (int)*local_100) && (*local_100 = *local_100 + -1, *local_100 == 0)) {
    (**(code **)(local_100[1] + 0x30))();
  }
  if ((-1 < (int)*local_f8) && (*local_f8 = *local_f8 + -1, *local_f8 == 0)) {
    (**(code **)(local_f8[1] + 0x30))();
  }
  if ((-1 < (int)*local_f0) && (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
    (**(code **)(local_f0[1] + 0x30))();
  }
  if ((-1 < (int)*local_e8) && (*local_e8 = *local_e8 + -1, *local_e8 == 0)) {
    (**(code **)(local_e8[1] + 0x30))();
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if ((-1 < *(int *)pcVar17) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
  if ((-1 < (int)*local_c8) && (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
    (**(code **)(local_c8[1] + 0x30))();
  }
  if ((-1 < (int)*local_c0) && (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
    (**(code **)(local_c0[1] + 0x30))();
  }
  plVar4 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar3;
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  return 0;
}

